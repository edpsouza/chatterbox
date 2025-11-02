# Chatterbox

**End-to-End Encrypted Chat System in Go**

<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/go-1.21-blue?logo=go" /></a>
  <a href="https://github.com/edpsouza/chatterbox/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-green" /></a>
  <a href="https://github.com/edpsouza/chatterbox/actions"><img src="https://github.com/edpsouza/chatterbox/actions/workflows/ci.yml/badge.svg?branch=main" alt="CI Status" /></a>
  <img src="https://img.shields.io/badge/security-E2EE-important?logo=lock" />
  <img src="https://img.shields.io/badge/backend-Go-blue?logo=go" />
  <img src="https://img.shields.io/badge/client-Go-blue?logo=go" />
</p>

---

## Overview

Chatterbox is a privacy-first, modular chat system built in Go.
It uses end-to-end encryption (Curve25519/X25519) so only intended recipients can read messages.
The backend never sees plaintext—only encrypted data.

---

## Features

- End-to-end encrypted messaging (ECC)
- Modular Go backend (WebSocket, SQLite)
- Public key exchange and lookup
- Secure registration and authentication
- Recipient-only message routing
- Message persistence (ciphertext only)
- Seamless chat history loading (previous messages appear when you rejoin a conversation)
- Easy local multi-user testing

---

## Architecture

```
chatterbox/
  cmd/                # Backend entrypoint
  internal/           # Backend modules
```

---

## Getting Started

### Backend

```sh
cd chatterbox
go mod tidy
go run ./cmd/main.go
```

---

## Security

- All messages encrypted client-side (Curve25519)
- Backend stores only ciphertext
- Public keys exchanged via backend; private keys never leave client

---

## Usage Example: Seamless Chat History

When you start a chat with a user, previous messages are automatically loaded and displayed after authentication:

```
Authenticated. Start chatting! Type 'exit' to leave.
---- Chat History ----
[2024-06-10T12:34:56Z] alice -> bob: Hey Bob, are you there?
[2024-06-10T12:35:10Z] bob -> alice: Hi Alice! Yes, I'm here.
----------------------
>
```

You can pick up the conversation right where you left off!

---

## Troubleshooting

- **JWT secret not set:**
  If you see "JWT secret not set" errors, set the `JWT_SECRET` environment variable before running the backend.

- **Database not found:**
  If you get "database not found," ensure you have run the backend at least once to create `chatterbox.db`.

- **Port conflicts:**
  If the server fails to start, make sure port `8080` is available or change it in your `.env` file.

- **Client cannot connect:**
  Ensure the backend is running and accessible at the configured address.

---

## Testing & CI

Automated tests are run on every push and pull request via GitHub Actions (see CI badge above).
To run tests locally:

```sh
go test ./...
```

- Unit and integration tests cover password hashing, registration/login, message persistence, and history.
- CI ensures code quality and reliability for every commit.

---

## Roadmap

**Near-Term**
- [ ] Group chat support
- [x] Message history retrieval
- [ ] User presence/status
- [ ] Improved CLI error handling
- [ ] Automated tests & CI

**Mid-Term**
- [ ] Desktop/mobile GUI client
- [ ] File/media sharing
- [ ] Multi-device support
- [ ] Push notifications

**Long-Term**
- [ ] Forward secrecy (ratcheting protocols)
- [ ] Invite links/QR codes
- [ ] E2EE voice/video calls
- [ ] Advanced user profiles

---

## Contributing

Pull requests and issues are welcome!
See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).

---

## Maintainers

- [@edpsouza](https://github.com/edpsouza)

---

## Acknowledgements

- [Go](https://golang.org)
- [gorilla/websocket](https://github.com/gorilla/websocket)
- [x/crypto/nacl/box](https://pkg.go.dev/golang.org/x/crypto/nacl/box)
