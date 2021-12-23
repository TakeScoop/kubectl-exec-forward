package command

const sensitiveAsterisks = "********"

func sensitiveFunc(showSensitive bool) func(string) string {
	return func(s string) string {
		if showSensitive {
			return s
		}

		return sensitiveAsterisks
	}
}
