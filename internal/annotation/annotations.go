package annotation

const (
	// Args is the annotation key used to store arguments to pass to the commands.
	Args = "exec-forward.pod.kubernetes.io/args"
	// PreConnect is the annotation key name used to store commands run before establishing a portforward connection.
	PreConnect = "exec-forward.pod.kubernetes.io/pre-connect"
	// PostConnect is the annotation key name used to store commands run after establishing a portforward connection.
	PostConnect = "exec-forward.pod.kubernetes.io/post-connect"
	// Command is the annotation key name used to store the main command to run after the post-connect hook has been run.
	Command = "exec-forward.pod.kubernetes.io/command"
)
