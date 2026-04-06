.PHONY: build test lint format check quality clean fe-dev be-dev fe-build be-build fe-test be-test fe-lint be-lint harvest

# === Frontend (Next.js / TypeScript) ===
fe-install:
	cd frontend && pnpm install

fe-build:
	cd frontend && pnpm build

fe-test:
	cd frontend && pnpm test

fe-lint:
	cd frontend && pnpm lint

fe-format:
	cd frontend && pnpm format

fe-typecheck:
	cd frontend && pnpm typecheck

fe-dev:
	cd frontend && pnpm dev

# === Backend (Go) ===
be-build:
	cd backend && go build -trimpath -ldflags "-s -w" -o bin/server ./cmd/server/

be-test:
	cd backend && go test -v -race -count=1 -coverprofile=coverage.out ./...

be-lint:
	cd backend && golangci-lint run ./...

be-format:
	cd backend && gofumpt -w .

be-tidy:
	cd backend && go mod tidy

be-dev:
	cd backend && go run ./cmd/server/

# === Combined ===
build: fe-build be-build
	@echo "Build complete."

test: fe-test be-test
	@echo "All tests passed."

lint: fe-lint be-lint
	@echo "Lint clean."

format: fe-format be-format
	@echo "Format complete."

check: format lint fe-typecheck test build
	@echo "All checks passed."

quality:
	@echo "=== Quality Gate ==="
	@test -f LICENSE || { echo "ERROR: LICENSE missing. Fix: add MIT LICENSE file"; exit 1; }
	@! grep -rn "TODO\|FIXME\|HACK\|console\.log" frontend/src/ 2>/dev/null | grep -v "node_modules" || { echo "ERROR: debug output or TODO found in frontend. Fix: remove before ship"; exit 1; }
	@! grep -rn "TODO\|FIXME\|HACK\|fmt\.Print" backend/internal/ backend/cmd/ 2>/dev/null || { echo "ERROR: debug output or TODO found in backend. Fix: remove before ship"; exit 1; }
	@! grep -rn "password=\|secret=\|api_key=\|sk-\|ghp_" frontend/src/ backend/ 2>/dev/null | grep -v '\$$' | grep -v "node_modules" | grep -v "_test.go" | grep -v ".example" || { echo "ERROR: hardcoded secrets. Fix: use env vars"; exit 1; }
	@test ! -f CLAUDE.md || [ $$(wc -l < CLAUDE.md) -le 50 ] || { echo "ERROR: CLAUDE.md is $$(wc -l < CLAUDE.md) lines (max 50)"; exit 1; }
	@echo "OK: automated quality checks passed"

clean:
	rm -rf frontend/.next frontend/dist frontend/coverage backend/bin backend/coverage.out

harvest:
	@mkdir -p docs
	@echo "# Harvest: agile-metrics-hub" > docs/harvest.md
	@echo "" >> docs/harvest.md
	@echo "## メトリクス" >> docs/harvest.md
	@echo "| 項目 | 値 |" >> docs/harvest.md
	@echo "|------|-----|" >> docs/harvest.md
	@echo "| コミット数 | $$(git log --oneline --no-merges | wc -l) |" >> docs/harvest.md
	@echo "| ADR数 | $$(ls docs/adr/*.md 2>/dev/null | wc -l) |" >> docs/harvest.md
	@echo "| CLAUDE.md行数 | $$(wc -l < CLAUDE.md 2>/dev/null || echo 0) |" >> docs/harvest.md
	@echo "" >> docs/harvest.md
	@echo "Harvest report: docs/harvest.md"
