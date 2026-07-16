# diary

James Yusuke が書く、Go・Web・設計についての個人技術ブログです。記事はこのGitHubリポジトリでMarkdownとして管理します。

## 必要なもの

- Go 1.26 以降

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
