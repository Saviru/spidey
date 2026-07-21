package core

type Directories struct {
	PublicDir string `json:"publicDir"`
	OutputDir string `json:"outputDir"`
}

type Config struct {
	Port        int         `json:"port"`
	Directories Directories `json:"directories"`
}
