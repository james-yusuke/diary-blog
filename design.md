# Design — diary

## Purpose

実務エンジニアが新着の技術記事をすぐに読み、必要に応じてCVE・セキュリティを検索やタグから辿れる個人ブログ。

## System

- Genre: modern-minimal
- Theme: Coral — warm-white paper, restrained coral signal, dark ink actions
- Display: Manrope 600–700
- Body: IBM Plex Sans 400–600
- Mono: JetBrains Mono 400–600
- Spacing: `assets/tokens.css` の4px基準トークンだけを使う
- Motion: 原則静的。操作フィードバックだけを120–220msの`transform`/`opacity`で扱い、reduced motionでは即時化する
- Color mode: 初期値はOS設定に追従する。ヘッダーの単一ボタンで自動・ライト・ダークを循環し、最初の操作では現在の見た目と反対のテーマへ切り替える。明示選択はブラウザに保存する

## Page families

- Home / tag / search: Index-First。記事ストリームとトピック一覧を優先し、宣伝的なヒーローは置かない
- Post: Long Document。本文、メタデータ、Mermaid、コードを読みやすい一列の流れにする
- About / 404: 短い本文中心。カードを増やさない

## Shared chrome

- Navigation: N9 edge-aligned minimal。記事一覧・Security・About、検索トリガー、テーマ切替だけを置く
- Footer: Ft2 inline rule。短いクレジットとGitHub / RSSリンクを一行で閉じる
- CTA: 塗りつぶしではなく下線付きテキストリンクを基本とし、検索など操作が必要な場面だけ暗いインクのボタンを使う
- Accent: coralはキーボードフォーカス、現在地、短い注記に限定する。大面積には使わない

## Constraints

- 既存のURL、検索クエリ、RSS、コンテンツ形式、Zenn連携を維持する
- `assets/tokens.css` を唯一のトークン定義とし、他のスタイルはトークン参照だけを使う
- 320 / 375 / 414 / 768pxで横スクロールと複数行のクリック可能ラベルを発生させない
