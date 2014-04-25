package main

func init() {
	register("init", initCmd, "Initializes a browserflood project in the current directory.")
}

func initCmd() error {
	return InitProject()
}
