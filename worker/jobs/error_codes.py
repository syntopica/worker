"""Every error code an attempt may settle with (spec 9); ``http_###`` is added by pattern."""

ERROR_CODES = frozenset(
    {
        "bad_response",
        "cancelled",
        "drain_failed",
        "executor_error",
        "host_state_unreadable",
        "input_denied",
        "input_missing",
        "input_too_large",
        "inputs_not_allowed",
        "lease_lost",
        "memory_pressure",
        "no_output",
        "node_error",
        "node_shutdown",
        "on_battery",
        "privacy_refused",
        "quota_wall",
        "rate_limited",
        "runner_auth",
        "runner_failed",
        "runner_unavailable",
        "schema_violation",
        "timeout",
        "transport_error",
        "unknown_model",
        "user_active",
    }
)
