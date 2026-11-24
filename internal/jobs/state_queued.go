package jobs

const (
	stateQueued         = "queued"
	stateLeased         = "leased"
	stateRunning        = "running"
	stateDraining       = "draining"
	stateSplitRequested = "split_requested"
	stateSuperseded     = "superseded"
	stateSucceeded      = "succeeded"
	stateFailed         = "failed"
	stateExpired        = "expired"
	stateCancelled      = "cancelled"
	stateUnackedExpired = "unacked_expired"

	kindInference  = "inference"
	kindTask       = "task"
	kindOpenRouter = "openrouter"

	executorOllama = "ollama"
	executorRunner = "runner"

	pinJudge    = "judge"
	pinOriginal = "original"

	producerShadow = "_shadow"
	producerJudge  = "_judge"

	codeExecutorError   = "executor_error"
	codeSchemaViolation = "schema_violation"
	codeQuotaWall       = "quota_wall"
	codeNotFound        = "not_found"
	codeBadInput        = "bad_input"
	codeBadSplitChild   = "bad_split_child"
	codeOutOfRange      = "out_of_range"

	keyProvider = "provider"
	keyModel    = "model"
	keyInputs   = "inputs"
	keyLabels   = "labels"
	keyMessages = "messages"
	keySchema   = "schema"
	keyProfile  = "profile"
	keyRunner   = "runner"
	keyInput    = "input"
	keyError    = "error"

	daySeconds  = 86400.0
	hourSeconds = 3600.0
)
