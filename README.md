# diary

James Yusuke が書く、Go・Web・設計についての個人技術ブログです。記事はこのGitHubリポジトリでMarkdownとして管理します。

## 必要なもの

- Go 1.23 系（ローカル開発とTinyGoビルドで共通）

依存関係はGo moduleで管理しています。初回だけ実行してください。

```sh
go mod download
```

## ローカル起動

```sh
make run
```

ブラウザで `http://localhost:8080` を開きます。ポートを変える場合は `ADDR=:3000 make run` のように指定できます。

## 記事を書く

1. `content/posts/TEMPLATE.md` を複製し、日付と内容に合う名前へ変更します。
2. front matter の `slug`、`title`、`summary`、`published_at`、`tags` を設定します。
3. `draft: true` の間は一覧・RSSには出ません。公開するときは削除するか `false` にします。
4. MarkdownをコミットしてGitHubへプッシュします。

記事の一覧、タグ、検索、RSSはfront matterから自動で組み立てられます。不正なメタデータやslugの重複は起動時にエラーになります。

## 検証

```sh
make test
make build
```

GitHub Actionsもmainブランチへのpushとpull requestで、templ生成・テスト・ビルドを実行します。

## Cloudflare Workers（TinyGo）

Cloudflare Workersでは実行時にローカルファイルを読めないため、Workerビルド時に `content/posts`、`config/site.yaml`、`assets/site.css` をGoソースへ変換してWasmに含めます。記事の正本は引き続きこのリポジトリのMarkdownです。

必要なもの：

- Go 1.23 系（TinyGo 0.38との互換性のため）
- TinyGo 0.38 以降
- Binaryen（`wasm-opt`。macOSでは `brew install binaryen`）
- Node.js と `npm install`（WranglerによるローカルWorker実行時のみ）

Worker用の成果物を生成します。デプロイは実行しません。

```sh
make GO=go1.23.6 worker-build
```

`build/app.wasm`、`build/worker.mjs`、`build/wasm_exec.js`、`build/runtime.mjs` がCloudflare Workers用の一式です。Wranglerを利用する場合は、同じGoツールチェーンを選んでから次を実行します。

```sh
GO=go1.23.6 npm run dev
```

`wrangler deploy` は `npm run deploy` として定義していますが、このリポジトリでは実行しません。
