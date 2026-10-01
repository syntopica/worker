import pytest

from worker.node.redact_credentials import redact_credentials


@pytest.mark.parametrize(
    ("text", "expected"),
    [
        ("Your code is 482913.", "Your code is [REDACTED]."),
        ("Tu código de verificación: 1234", "Tu código de verificación: [REDACTED]"),
        ("password: hunter22!", "password: [REDACTED]"),
        ("Contraseña = Abc12345", "Contraseña = [REDACTED]"),
        ("https://bob:s3cret@host/x", "https://bob:[REDACTED]@host/x"),
        ("key sk-ant-abcdefghijklmnopqrstuvwxyz0123", "key [REDACTED]"),
        ("https://x.com/reset?token=abcdefgh&u=1", "https://x.com/reset?token=[REDACTED]&u=1"),
        ("Authorization: Bearer abcdefghijklmnopqrstuvwxyz", "Authorization: Bearer [REDACTED]"),
        ("ghp_abcdefghijklmnopqrstuvwxyz0123456789", "[REDACTED]"),
        ("AKIAABCDEFGHIJKLMNOP", "[REDACTED]"),
        ("-----BEGIN RSA PRIVATE KEY-----\nMIIE\n-----END RSA PRIVATE KEY-----", "[REDACTED]"),
    ],
)
def test_credentials_are_redacted(text, expected):
    assert redact_credentials(text) == expected


def test_ordinary_text_is_untouched():
    text = "Factura 2024 por 35,49 EUR, cita el 10-01 a las 09:30."
    assert redact_credentials(text) == text
