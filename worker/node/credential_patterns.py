"""Patterns for credentials that must never reach a remote executor.

Each pattern's group ``secret`` (or the whole match when it has none) is
replaced. They err toward redacting: a postal code after "código" is a
smaller loss than a one-time code sent to a third party.
"""

import re

CREDENTIAL_PATTERNS = (
    re.compile(r"-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----"),
    re.compile(
        r"\b(?:sk-(?:ant-|proj-)?[A-Za-z0-9_-]{20,}|gh[pousr]_[A-Za-z0-9]{20,}"
        r"|github_pat_[A-Za-z0-9_]{20,}|glpat-[A-Za-z0-9_-]{20,}|xox[abprs]-[A-Za-z0-9-]{10,}"
        r"|AKIA[0-9A-Z]{16}|AIza[0-9A-Za-z_-]{35}|sk_live_[0-9A-Za-z]{16,}"
        r"|eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,})"
    ),
    re.compile(r"(?i)\bBearer\s+(?P<secret>[A-Za-z0-9._~+/=-]{16,})"),
    re.compile(r"(?i)://[^\s/:@]+:(?P<secret>[^\s/@]+)@"),
    re.compile(
        r"(?i)\b(?:password|passwd|pwd|contrase[ñn]a|clave|passcode|pin|token|secret|"
        r"api[ _-]?key|access[ _-]?key|client[ _-]?secret)\b[\"']?\s*(?:[:=]|is|es)\s*[\"']?"
        r"(?P<secret>[^\s\"',;<>&]{4,})"
    ),
    re.compile(
        r"(?i)\b(?:code|c[óo]digo|verification|verificaci[óo]n|otp|2fa|one[- ]time|"
        r"security|seguridad)\b[^\d\n]{0,40}?(?P<secret>\b\d{4,8}\b)"
    ),
    re.compile(
        r"(?i)[?&](?:token|key|code|sig|signature|auth|access_token|reset|otp)="
        r"(?P<secret>[^&\s\"'<>]{6,})"
    ),
)
