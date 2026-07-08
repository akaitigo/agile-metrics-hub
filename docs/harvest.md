# Harvest: agile-metrics-hub

> v1.0.0 MVP 振り返り (2026-04-06)

## メトリクス

| 項目 | 値 |
|------|-----|
| Issue (closed/total) | 6/6 |
| PR merged | 10 |
| テスト数 | 32 |
| CI失敗数 | 2（初期CI設定の修正。go.sum不在 + golangci-lint未インストール） |
| ADR数 | 4 |
| コミット数 (non-merge) | 11 |
| CLAUDE.md行数 | 38（上限50以内） |
| テンプレート実装率 | 92%（Layer-0: 8/8, Layer-1: 5/5, Layer-2: 2/3） |

## ハーネス適用状況

### Layer-0 (リポジトリ衛生)
- [x] CLAUDE.md（38行、50行以下）
- [x] Makefile（check/quality ターゲット）
- [x] LICENSE（MIT）
- [x] .gitignore
- [x] lefthook.yml（pre-commit: lint/format/test 並列）
- [x] startup.sh（ツール自動インストール + lefthook install）
- [x] ADR 4件
- [x] dependabot.yml + auto-merge

### Layer-1 (ツール強制)
- [x] .claude/settings.json（PreToolUse/PostToolUse/PreCompact/Stop hooks）
- [x] CI（.github/workflows/ci.yml — frontend + backend 並列）
- [x] lint設定（biome.json + golangci-lint）
- [x] format設定（biome format + gofumpt）
- [x] 型チェック（tsc --noEmit + Go コンパイラ）

### Layer-2 (プロセス)
- [x] Issue-PR 1:1対応（6 Issue → 6 PR + 4 review/ship PR）
- [x] model ラベル分類（haiku: 1, opus: 1, sonnet: 4）
- [ ] PRD 受け入れ条件が全 [x]（未チェック — MVPデモデータ段階のため）

## テンプレートへの改善提案

| # | 優先度 | 対象ファイル | 変更内容 | 根拠 |
|---|--------|-------------|---------|------|
| 1 | HIGH | idea-launch/SKILL.md | モノレポ対応の明示: frontend/backend 分割時のMakefile構造、hooks分離、CI設定のガイドを追加 | 今回のモノレポでCI/hooks設定に試行錯誤があった |
| 2 | HIGH | layer-1-language/go/Makefile | CI用に `golangci-lint-action` を使う旨をREADMEに記載 | CI上でgolangci-lintが未インストールでCI失敗。ローカルとCIの差異 |
| 3 | MEDIUM | idea-work/SKILL.md | `/idea-work` をスキップした場合のガイダンス: startup.sh実行、セキュリティチェック、破壊的検証が抜けるリスクを明記 | 今回idea-workをスキップしてこれらが全て抜けた |
| 4 | MEDIUM | layer-0-universal/startup.sh | モノレポ対応: backend/frontend サブディレクトリの自動検出 | 現状は単一言語前提。今回はカスタムstartup.shを書いた |
| 5 | LOW | idea-launch/SKILL.md | `gh label create` のコロン付きラベル名の対応 | IFS=: でパースするとラベル名のコロンも分割されて失敗した |

## 振り返り

### 良かった点
- **Adapter パターン**: PMToolAdapter インターフェースにより、ClickUp/Jira のAPI差分を完全に吸収。新規PMツール追加はファクトリー関数1行で登録可能
- **AES-256-GCM 暗号化**: APIキー保護が構造的に正しく実装でき、テストカバレッジ79%
- **レビューループの効果**: R2/R3でSSRF、Slowloris、エラー情報漏洩という重大な脆弱性を発見・修正。Ship前レビューは必須
- **テンプレート活用**: Layer-0/1/2 のテンプレートにより、CI/hooks/品質ゲートが初期からセットアップされ、開発速度に寄与
- **ADR**: 4件のADRにより設計判断が記録され、将来の「なぜこうなった？」を防止

### 改善点
- **`/idea-work` スキップ**: 直接実装に走った結果、startup.sh実行、Issueステータスラベル遷移、セキュリティチェック、破壊的検証が全て抜けた。ワークフローのスキップは品質低下に直結する
- **モノレポ対応**: テンプレートが単一言語前提のため、モノレポ（Go + TypeScript）ではMakefile/hooks/CI/startup.shを全てカスタム記述する必要があった
- **DB統合未完**: sqlcクエリ定義・マイグレーションを作成したが、ハンドラーとの配線が未完了（ADR #004 で明示的に先送り判断を記録）
- **CI初期設定**: go.sum不在、golangci-lint未インストールで2回CI失敗。ローカルとCIの差異をテンプレートで吸収すべき

## 次のPJへの申し送り

1. **ワークフローをスキップしない**: `/idea-work` → `/idea-ship` → `/idea-review-loop` → `/idea-harvest` の順序は品質を担保する構造。直接実装は短期的には速いが、セキュリティ・品質の穴を生む
2. **モノレポ時はMakefile設計を先に決める**: `make check` がルートから frontend/backend 両方を叩く構造を最初に設計する
3. **SSRF防止はアダプター生成時に検証する**: ユーザー入力URLを外部リクエストに使うパターンは必ずホワイトリスト検証を入れる
4. **http.Server は必ずタイムアウト設定する**: `http.ListenAndServe` は使わない。`http.Server` 構造体で Read/Write/Idle タイムアウトを明示する
5. **エラーメッセージをサニタイズする**: `err.Error()` をそのままクライアントに返さない。sentinel error でマッチして汎用メッセージに変換する
