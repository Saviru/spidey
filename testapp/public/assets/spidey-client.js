
document.addEventListener("DOMContentLoaded", () => {
	const islands = document.querySelectorAll("spidey-island");
	islands.forEach(async (island) => {
		const compName = island.getAttribute("data-component");
		if (compName) {
			try {
				const module = await import("/assets/components/" + compName + ".js");
				if (module.mount) {
					module.mount(island);
				}
			} catch (e) {
				console.error("Spidey: Failed to load island", compName, e);
			}
		}
	});
});
