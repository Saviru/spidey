package bundler

import (
	"crypto/rand"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/saviru/spidey/internal/config"
	"github.com/saviru/spidey/internal/parser"

	"github.com/evanw/esbuild/pkg/api"
)

func getModuleName(projectDir string) string {
	content, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
	if err == nil {
		for _, line := range strings.Split(string(content), "\n") {
			if strings.HasPrefix(line, "module ") {
				return strings.TrimSpace(strings.TrimPrefix(line, "module "))
			}
		}
	}
	return "spideyapp"
}

func buildNestedLayout(pagesDir, pagePath, appLayoutStr string) string {
	currentLayout := appLayoutStr

	relPath, err := filepath.Rel(pagesDir, pagePath)
	if err != nil {
		return currentLayout
	}

	dir := filepath.Dir(relPath)
	if dir == "." {
		return currentLayout
	}

	parts := strings.Split(filepath.ToSlash(dir), "/")
	currentDir := pagesDir

	for _, part := range parts {
		currentDir = filepath.Join(currentDir, part)
		layoutPath := filepath.Join(currentDir, "layout.spidey")
		if content, err := os.ReadFile(layoutPath); err == nil {
			currentLayout = strings.Replace(currentLayout, `{{template "content" .}}`, string(content), 1)
		}
	}

	return currentLayout
}

func setupGeneratedDirectory(projectDir string, templates embed.FS) error {
	genDir := filepath.Join(projectDir, "hub", "pages")

	os.RemoveAll(genDir)
	os.MkdirAll(genDir, 0755)

	// Read separate base.go template and inject it safely
	baseCodeBytes, err := templates.ReadFile("templates/base.go")
	if err != nil {
		return fmt.Errorf("could not read base.go template: %v", err)
	}
	baseCode := strings.Replace(string(baseCodeBytes), "//go:build ignore", "", 1)
	baseCode = strings.TrimSpace(baseCode) + "\n"

	return os.WriteFile(filepath.Join(genDir, "spidey_base.go"), []byte(baseCode), 0644)
}

func prepareAppLayout(projectDir string, liveReloadPort string, usesAOT bool) string {
	appLayoutStr := ""
	appLayoutPath := filepath.Join(projectDir, "app.spidey")
	if appLayoutBytes, err := os.ReadFile(appLayoutPath); err == nil {
		appLayoutStr = string(appLayoutBytes)

		// Inject scoped CSS stylesheet
		cssLink := `<link rel="stylesheet" href="/assets/spidey.css">`
		if !strings.Contains(appLayoutStr, cssLink) {
			if strings.Contains(appLayoutStr, "</head>") {
				appLayoutStr = strings.Replace(appLayoutStr, "</head>", cssLink+"\n</head>", 1)
			} else {
				appLayoutStr = cssLink + "\n" + appLayoutStr
			}
		}

		// Inject Client Bootstrapper for Islands
		clientScript := `<script src="/assets/spidey-client.js?v=2" type="module"></script>`
		if !strings.Contains(appLayoutStr, clientScript) {
			if strings.Contains(appLayoutStr, "</body>") {
				appLayoutStr = strings.Replace(appLayoutStr, "</body>", clientScript+"\n</body>", 1)
			} else {
				appLayoutStr += "\n" + clientScript
			}
		}

		// Inject Livereload if in dev mode
		if liveReloadPort != "" {
			script := fmt.Sprintf(`<script>const evtSource = new EventSource("http://localhost:%s/livereload");evtSource.onmessage = function(e) { if(e.data === "reload") { setTimeout(() => window.location.reload(), 100); } };</script>`, liveReloadPort)
			if strings.Contains(appLayoutStr, "</body>") {
				appLayoutStr = strings.Replace(appLayoutStr, "</body>", script+"\n</body>", 1)
			} else {
				appLayoutStr += "\n" + script
			}
		}

		// inhect aot js file
		if usesAOT {
			aotScript := `<script src="/assets/spidey-aot.js"></script>`
			if !strings.Contains(appLayoutStr, aotScript) {
				if strings.Contains(appLayoutStr, "</body>") {
					appLayoutStr = strings.Replace(appLayoutStr, "</body>", aotScript+"\n</body>", 1)
				} else {
					appLayoutStr += "\n" + aotScript
				}
			}
		}
	}
	return appLayoutStr
}

func processComponents(projectDir string) (string, *strings.Builder) {
	componentsDir := filepath.Join(projectDir, "components")
	pagesDir := filepath.Join(projectDir, "pages")
	var componentsBuilder strings.Builder
	var globalStyles strings.Builder

	processFile := func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".spidey") {
			content, _ := os.ReadFile(path)
			name := strings.TrimSuffix(filepath.Base(path), ".spidey")

			parsed, err := parser.Parse(name, string(content))
			if err == nil {
				if parsed.Styles != "" {
					globalStyles.WriteString(parsed.Styles + "\n")
				}
				// Wrap HTML in a Go template define block
				componentsBuilder.WriteString(fmt.Sprintf("\n{{define \"%s\"}}\n%s\n{{end}}\n", name, parsed.HTML))
			}
		}
		return nil
	}

	// Process global components
	filepath.WalkDir(componentsDir, processFile)

	// Process colocated components in pages directory (ignored by router)
	filepath.WalkDir(pagesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(pagesDir, path)
		parts := strings.Split(filepath.ToSlash(relPath), "/")
		isPrivate := false
		for _, part := range parts {
			if strings.HasPrefix(part, "_") {
				isPrivate = true
				break
			}
		}

		if isPrivate {
			return processFile(path, d, err)
		}
		return nil
	})

	return componentsBuilder.String(), &globalStyles
}

func generateAPIRoutes(projectDir string) (string, bool) {
	apiDir := filepath.Join(projectDir, "api")
	var apiRoutesBuilder strings.Builder
	hasApiRoutes := false

	if _, err := os.Stat(apiDir); err == nil {
		filepath.WalkDir(apiDir, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(path, ".go") && filepath.Base(path) != "main.go" {
				contentBytes, _ := os.ReadFile(path)
				content := string(contentBytes)

				re := regexp.MustCompile(`((?:[ \t]*//spidey:[^\n]*\n)+)func\s+([A-Za-z0-9_]+)`)
				matches := re.FindAllStringSubmatch(content, -1)

				for _, match := range matches {
					if len(match) == 3 {
						commentsBlock := match[1]
						funcName := match[2]

						var middlewares []string
						var method, routePath string

						lines := strings.Split(commentsBlock, "\n")
						for _, line := range lines {
							line = strings.TrimSpace(line)
							if strings.HasPrefix(line, "//spidey:middleware") {
								parts := strings.Fields(line)
								if len(parts) >= 2 {
									middlewares = append(middlewares, parts[1])
								}
							} else if strings.HasPrefix(line, "//spidey:route") {
								parts := strings.Fields(line)
								if len(parts) >= 3 {
									method = parts[1]
									routePath = parts[2]
								}
							}
						}

						if method != "" && routePath != "" {
							pathRe := regexp.MustCompile(`\[([^\]]+)\]`)
							routePath = pathRe.ReplaceAllString(routePath, `{$1}`)

							if len(middlewares) > 0 {
								mwArgs := make([]string, len(middlewares))
								for i, mw := range middlewares {
									mwArgs[i] = "api." + mw
								}
								apiRoutesBuilder.WriteString(fmt.Sprintf("\tapp.Group(\"\", %s).Handle(\"%s\", \"%s\", api.%s)\n", strings.Join(mwArgs, ", "), method, routePath, funcName))
							} else {
								apiRoutesBuilder.WriteString(fmt.Sprintf("\tapp.Handle(\"%s\", \"%s\", api.%s)\n", method, routePath, funcName))
							}
							hasApiRoutes = true
						}
					}
				}
			}
			return nil
		})
	}
	return apiRoutesBuilder.String(), hasApiRoutes
}

func transpilePages(projectDir string, appLayoutStr string, componentsStr string, globalStyles *strings.Builder, apiRoutesStr string, hasApiRoutes bool, aotJSBuffer *strings.Builder) error {
	pagesDir := filepath.Join(projectDir, "pages")
	genDir := filepath.Join(projectDir, "hub", "pages")

	var routesBuilder strings.Builder
	routesBuilder.WriteString("// Code generated by Spidey. DO NOT EDIT.\n")
	routesBuilder.WriteString("package pages\n\n")
	routesBuilder.WriteString("import (\n")
	routesBuilder.WriteString("\t\"github.com/saviru/spidey/pkg/core\"\n")
	if hasApiRoutes {
		routesBuilder.WriteString(fmt.Sprintf("\tapi \"%s/api\"\n", getModuleName(projectDir)))
	}
	routesBuilder.WriteString(")\n\n")
	routesBuilder.WriteString("func RegisterRoutes(app *core.App) {\n")

	// Transpile .spidey files
	err := filepath.WalkDir(pagesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip files or folders starting with "_"
		if strings.HasPrefix(d.Name(), "_") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			if strings.HasSuffix(path, ".jsx") {
				fmt.Println("Error: JSX is not supported in Spidey pages.")
				return nil
			} else if strings.HasSuffix(path, ".js") {
				fmt.Printf("Warning: Please avoid using .js files directly (%s). Use the native .spidey format in pages.\n", filepath.Base(path))
			}
		}

		if !d.IsDir() && strings.HasSuffix(path, ".spidey") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if filepath.Base(path) == "layout.spidey" {
				return nil
			}

			relPath, _ := filepath.Rel(pagesDir, path)
			relPath = filepath.ToSlash(relPath)
			componentName := strings.TrimSuffix(relPath, ".spidey")

			compiledHTML := CompileAOT(string(content), aotJSBuffer)
			pageLayoutStr := buildNestedLayout(pagesDir, path, appLayoutStr)

			goCode, err := parser.TranspileToGo(componentName, compiledHTML, pageLayoutStr, componentsStr, globalStyles)
			if err != nil {
				return err
			}

			safeName := strings.ReplaceAll(componentName, "/", "_")
			safeName = strings.ReplaceAll(safeName, "[", "")
			safeName = strings.ReplaceAll(safeName, "]", "")
			genPath := filepath.Join(genDir, safeName+"_spidey.go")
			os.WriteFile(genPath, []byte(goCode), 0644)

			routePath := "/" + componentName
			if strings.HasSuffix(routePath, "/index") {
				routePath = strings.TrimSuffix(routePath, "index")
			}
			if routePath != "/" && strings.HasSuffix(routePath, "/") {
				routePath = strings.TrimSuffix(routePath, "/")
			}

			re := regexp.MustCompile(`\[([^\]]+)\]`)
			routePath = re.ReplaceAllString(routePath, `{$1}`)

			routesBuilder.WriteString(fmt.Sprintf("\tapp.GET(\"%s\", func(c *core.Context) {\n", routePath))

			matches := re.FindAllStringSubmatch(componentName, -1)
			if len(matches) > 0 {
				routesBuilder.WriteString("\t\tdata := map[string]interface{}{\n")
				for _, match := range matches {
					param := match[1]
					exportName := strings.ToUpper(param[0:1]) + param[1:]
					routesBuilder.WriteString(fmt.Sprintf("\t\t\t\"%s\": c.Param(\"%s\"),\n", exportName, param))
				}
				routesBuilder.WriteString("\t\t}\n")
				routesBuilder.WriteString(fmt.Sprintf("\t\thtml, err := Render(\"%s\", data)\n", componentName))
			} else {
				routesBuilder.WriteString(fmt.Sprintf("\t\thtml, err := Render(\"%s\", nil)\n", componentName))
			}

			routesBuilder.WriteString("\t\tif err != nil {\n\t\t\tc.Send(\"<h1 style='color:red;'>Template Error</h1><pre>\" + err.Error() + \"</pre>\")\n\t\t\treturn\n\t\t}\n")
			routesBuilder.WriteString("\t\tc.Send(html)\n")
			routesBuilder.WriteString("\t})\n")
		}
		return nil
	})

	if hasApiRoutes {
		routesBuilder.WriteString("\n\t// API Routes\n")
		routesBuilder.WriteString(apiRoutesStr)
	}

	routesBuilder.WriteString("}\n")
	os.WriteFile(filepath.Join(genDir, "routes.go"), []byte(routesBuilder.String()), 0644)

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("transpilation failed: %v", err)
	}
	return nil
}

func bundleFrontendAssets(projectDir string, globalStyles *strings.Builder, cfg *config.Config) error {
	componentsDir := filepath.Join(projectDir, "components")

	// Output global styles
	os.MkdirAll(filepath.Join(projectDir, cfg.Directories.PublicDir, "assets"), 0755)
	os.WriteFile(filepath.Join(projectDir, cfg.Directories.PublicDir, "assets", "spidey.css"), []byte(globalStyles.String()), 0644)

	// Write spidey-client.js bootstrapper
	clientCode := `
document.addEventListener("DOMContentLoaded", () => {
	// Inject default built-in transitions
	const style = document.createElement('style');
	style.textContent = ` + "`" + `
		/* Disable default mix-blend-mode which causes text vibration during crossfades */
		::view-transition-image-pair(spidey-fade), ::view-transition-image-pair(spidey-slide-up), 
		::view-transition-image-pair(spidey-slide-down), ::view-transition-image-pair(spidey-scale) {
			isolation: auto;
		}
		::view-transition-old(spidey-fade), ::view-transition-new(spidey-fade),
		::view-transition-old(spidey-slide-up), ::view-transition-new(spidey-slide-up),
		::view-transition-old(spidey-slide-down), ::view-transition-new(spidey-slide-down),
		::view-transition-old(spidey-scale), ::view-transition-new(spidey-scale) {
			mix-blend-mode: normal;
		}

		::view-transition-old(spidey-fade) { animation: fade-out 0.8s linear both; }
		::view-transition-new(spidey-fade) { animation: fade-in 0.8s linear both; }
		@keyframes fade-in { from { opacity: 0; } to { opacity: 1; } }
		@keyframes fade-out { from { opacity: 1; } to { opacity: 0; } }

		::view-transition-old(spidey-slide-up) { animation: slide-up-out 0.8s ease-in both; }
		::view-transition-new(spidey-slide-up) { animation: slide-up-in 0.8s ease-out both; }
		@keyframes slide-up-in { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }
		@keyframes slide-up-out { from { opacity: 1; transform: translateY(0); } to { opacity: 0; transform: translateY(-20px); } }
		
		::view-transition-old(spidey-slide-down) { animation: slide-down-out 0.8s ease-in both; }
		::view-transition-new(spidey-slide-down) { animation: slide-down-in 0.8s ease-out both; }
		@keyframes slide-down-in { from { opacity: 0; transform: translateY(-20px); } to { opacity: 1; transform: translateY(0); } }
		@keyframes slide-down-out { from { opacity: 1; transform: translateY(0); } to { opacity: 0; transform: translateY(20px); } }

		::view-transition-old(spidey-scale) { animation: scale-out 0.8s ease-in both; }
		::view-transition-new(spidey-scale) { animation: scale-in 0.8s ease-out both; }
		@keyframes scale-in { from { opacity: 0; transform: scale(0.95); } to { opacity: 1; transform: scale(1); } }
		@keyframes scale-out { from { opacity: 1; transform: scale(1); } to { opacity: 0; transform: scale(1.05); } }
	` + "`" + `;
	document.head.appendChild(style);

	// Islands
	const islands = document.querySelectorAll("spidey-island");
	islands.forEach(async (island) => {
		const compName = island.getAttribute("data-component");
		if (compName) {
			try {
				const module = await import("/assets/components/" + compName + ".js");
				if (module.mount) module.mount(island);
			} catch (e) {
				console.error("Spidey: Failed to load island", compName, e);
			}
		}
	});

	// SPIDEY S-TAGS ENGINE
	window.spideyProcessElement = function(el) {
		if (el._spidey_processed) return;
		el._spidey_processed = true;

		let isPost = el.hasAttribute("s-post");
		let isGet = el.hasAttribute("s-get");
		if (!isPost && !isGet) return;

		let trigger = el.getAttribute("s-trigger");
		if (!trigger) {
			trigger = isPost ? "submit" : "click";
		}

		let url = isPost ? el.getAttribute("s-post") : el.getAttribute("s-get");
		let triggerParts = trigger.split(" ");
		let eventName = triggerParts[0];
		let delay = 0;
		let every = 0;

		for (let part of triggerParts.slice(1)) {
			if (part.startsWith("delay:")) {
				delay = parseInt(part.split(":")[1].replace(/[^0-9]/g, '')) || 0;
			} else if (part.startsWith("every:")) {
				every = parseInt(part.split(":")[1].replace(/[^0-9]/g, '')) || 0;
			}
		}

		const execute = async (e) => {
			if (e && typeof e.preventDefault === 'function') e.preventDefault();
			
			const targetSelector = el.getAttribute("s-target");
			const swapStyle = el.getAttribute("s-swap") || "innerHTML";
			
			try {
				let options = {};
				if (isPost) {
					options.method = "POST";
					if (el.tagName === "FORM") {
						options.body = new FormData(el);
					} else {
						const form = el.closest("form");
						if (form) options.body = new FormData(form);
					}
				}
				
				// Handle URL parameters (e.g. for search inputs on s-get)
				let finalUrl = url;
				if (isGet && el.tagName === "INPUT" && el.name) {
					const paramChar = finalUrl.includes("?") ? "&" : "?";
					finalUrl += paramChar + encodeURIComponent(el.name) + "=" + encodeURIComponent(el.value);
				}

				const response = await fetch(finalUrl, options);
				const html = await response.text();
				
				if (targetSelector) {
					const targetEl = document.querySelector(targetSelector);
					if (targetEl) {
						const applySwap = () => {
							if (swapStyle === "outerHTML") {
								targetEl.outerHTML = html;
							} else {
								targetEl.innerHTML = html;
							}
							document.querySelectorAll("[s-get], [s-post]").forEach(window.spideyProcessElement);
						};

						const transitionName = el.getAttribute("@transition") || el.getAttribute("s-transition");
						if (transitionName && document.startViewTransition) {
							// Determine the view-transition-name to use
							let activeName = transitionName;
							if (activeName === "true" || activeName === "") {
								activeName = "spidey-fade"; // default smooth fade
							}
							
							targetEl.style.viewTransitionName = activeName;
							
							// Force browser to recalculate styles so the transition name is captured
							void targetEl.offsetWidth;
							
							const transition = document.startViewTransition(() => {
								applySwap();
							});
							
							transition.finished.finally(() => {
								targetEl.style.viewTransitionName = "";
							});
						} else {
							applySwap();
						}
					}
				}
			} catch (err) {
				console.error("Spidey Engine: request failed", err);
			}
		};

		if (eventName === "intersect") {
			let observer = new IntersectionObserver((entries) => {
				if (entries[0].isIntersecting) {
					execute();
				}
			});
			observer.observe(el);
		} else if (eventName === "every") {
			execute();
			setInterval(execute, every || parseInt(triggerParts[0].split(":")[1].replace(/[^0-9]/g, '')) || 1000);
		} else {
			let timeout;
			el.addEventListener(eventName, (e) => {
				if (delay > 0) {
					clearTimeout(timeout);
					timeout = setTimeout(() => execute(e), delay);
				} else {
					execute(e);
				}
			});
		}
	};

	// Initialize all elements on load
	document.querySelectorAll("[s-get], [s-post]").forEach(window.spideyProcessElement);
});
`

	os.WriteFile(filepath.Join(projectDir, cfg.Directories.PublicDir, "assets", "spidey-client.js"), []byte(clientCode), 0644)

	// esbuild components for islands
	var jsEntries []string
	filepath.WalkDir(componentsDir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if strings.HasSuffix(path, ".jsx") {
				fmt.Println("Error: JSX is not supported in Spidey.")
			} else if strings.HasSuffix(path, ".js") {
				fmt.Printf("Warning: Please avoid using .js files directly (%s). Use the native .spidey format in components.\n", filepath.Base(path))
				jsEntries = append(jsEntries, path)
			}
		}
		return nil
	})

	if len(jsEntries) > 0 {
		fmt.Println("Spidey: Bundling frontend islands...")
		api.Build(api.BuildOptions{
			EntryPoints:       jsEntries,
			Outdir:            filepath.Join(projectDir, cfg.Directories.PublicDir, "assets", "components"),
			Bundle:            true,
			MinifyWhitespace:  true,
			MinifyIdentifiers: true,
			MinifySyntax:      true,
			Write:             true,
			Format:            api.FormatESModule,
			Splitting:         true,
		})
	}

	return nil
}

func ProcessPages(projectDir string, templates embed.FS, liveReloadPort string, cfg *config.Config) error {
	err := setupGeneratedDirectory(projectDir, templates)
	if err != nil {
		return err
	}

	usesAOT := hasAOTActions(projectDir)

	appLayoutStr := prepareAppLayout(projectDir, liveReloadPort, usesAOT)
	componentsStr, globalStyles := processComponents(projectDir)
	apiRoutesStr, hasApiRoutes := generateAPIRoutes(projectDir)

	var aotJSBuffer strings.Builder

	err = transpilePages(projectDir, appLayoutStr, componentsStr, globalStyles, apiRoutesStr, hasApiRoutes, &aotJSBuffer)
	if err != nil {
		return err
	}
	if aotJSBuffer.Len() > 0 {
		finalJS := "document.addEventListener('DOMContentLoaded', () => {\n" + aotJSBuffer.String() + "\n});"
		os.WriteFile(filepath.Join(projectDir, cfg.Directories.PublicDir, "assets", "spidey-aot.js"), []byte(finalJS), 0644)
	}

	return bundleFrontendAssets(projectDir, globalStyles, cfg)
}

func generateAOTID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("s_%x", b) //8-character ID
}

func CompileAOT(htmlContent string, aotJSBuffer *strings.Builder) string {
	re := regexp.MustCompile(`@([a-z]+)="([^"]+)"`)
	compiledHTML := re.ReplaceAllStringFunc(htmlContent, func(match string) string {
		matches := re.FindStringSubmatch(match)
		eventName := matches[1]
		expression := matches[2]
		
		// Ignore @transition as it is used for View Transitions, not AOT events
		if eventName == "transition" {
			return match
		}

		uniqueID := generateAOTID()
		jsCode := fmt.Sprintf(`
	const el_%s = document.getElementById('%s');
	if (el_%s) {
		el_%s.addEventListener('%s', (e) => { %s });
	}
`, uniqueID, uniqueID, uniqueID, uniqueID, eventName, expression)

		aotJSBuffer.WriteString(jsCode)
		// Replace @click=".." with id="s-1234"
		return fmt.Sprintf(`id="%s"`, uniqueID)
	})
	return compiledHTML
}

func hasAOTActions(projectDir string) bool {
	hasAOT := false
	checkDir := func(dir string) {
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(path, ".spidey") {
				content, _ := os.ReadFile(path)
				matches := regexp.MustCompile(`@([a-z]+)="[^"]+"`).FindAllStringSubmatch(string(content), -1)
				for _, match := range matches {
					if match[1] != "transition" {
						hasAOT = true
						break
					}
				}
			}
			return nil
		})
	}

	checkDir(filepath.Join(projectDir, "pages"))
	checkDir(filepath.Join(projectDir, "components"))
	return hasAOT
}
