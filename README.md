blogsync
=======

[![Build Status](https://github.com/x-motemen/blogsync/workflows/test/badge.svg?branch=master)][actions]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![PkgGoDev](https://pkg.go.dev/badge/github.com/x-motemen/blogsync)](PkgGoDev)

[actions]: https://github.com/x-motemen/blogsync/actions?workflow=test
[coveralls]: https://coveralls.io/r/x-motemen/blogsync?branch=master
[license]: https://github.com/x-motemen/blogsync/blob/master/LICENSE
[PkgGoDev]: https://pkg.go.dev/github.com/x-motemen/blogsync

## Description

はてなBlog用のCLIクライアントです

## Installation

```console
% brew install Songmu/tap/blogsync
```

https://github.com/x-motemen/blogsync/releases から実行ファイルを直接取得できます。

ソースコードから最新版を使いたい場合は `go install github.com/x-motemen/blogsync@latest` してください。

## Usage

### Configuration

まず初めに設定ファイルを準備します。設定ファイルには以下の2種類があります。

- ローカル設定: `./blogsync.yaml`
- グローバル設定: `~/.config/blogsync/config.yaml`

両方存在する場合、設定の内容はマージされますが、ローカル設定の内容が優先されます。

設定ファイルの内容は以下のようなYAMLです。

```yaml
motemen.hatenablog.com:
  username: motemen
  password: <API KEY>
default:
  local_root: /Users/motemen/Dropbox/Blog
```

各項目の意味は次のとおり:

- キー（例: `motemen.hatenablog.com` ）: ブログID (blogID)。はてなブログのダッシュボードからブログの設定画面などを開いたとき、URL に含まれる文字列です。技術的には AtomPub API における「ブログID」になります。独自ドメインを利用していない場合は配信ドメインと一致します。
  - "default" という名前のキーは特別で、すべてのブログの項目のデフォルト値として扱われます。
- `<blog>.username`: そのブログに投稿するはてなユーザの ID。
- `<blog>.password`: そのブログに投稿するための API キー。はてなユーザのパスワードではありません。ブログの詳細設定画面 の「APIキー」で確認できます。
- `<blog>.local_root`: ブログのエントリを格納するパスのルート。
    - `$local_root/$blogID/` 配下にエントリが格納されます。`omit_domain` 設定がされている場合はブログIDは含まれず、local\_root直下にエントリーが格納されます
- `<blog>.omit_domain`: ブログエントリを格納するパスにブログIDを含めません。
- `<blog>.owner`: 編集対象のブログオーナーが自身とは別のユーザーの場合、ブログオーナーを個別に設定できます。
- `<blog>.entry_directory`: ブログエントリを格納するディレクトリ名を指定します。デフォルトは「/entry/」です。はてなブログで記事を配信するディレクトリを変更している場合に設定します。

#### 環境変数による設定

以下の環境変数により、ユーザーIDとパスワードを設定できます。環境変数の値は設定ファイルの内容より優先されます。

- `BLOGSYNC_USERNAME`: デフォルトのはてなユーザーID
- `BLOGSYNC_PASSWORD`: デフォルトのAPIキー

ただし、これらの環境変数はデフォルトのユーザーIDとAPIキーを設定するものなので、ブログ毎にユーザーIDとAPIキーが設定されている場合、これらの環境変数は無視されることに注意してください。この挙動は将来的に変更する可能性があります。

#### ブログオーナーが自身とは別の場合の設定

複数人で編集するブログなどで、編集者とブログのオーナーが別ユーザーの場合は下記のように設定できます。

```yaml
example.hatenablog.com:
  username: sample
  password: <API KEY>
  owner: <OWNER>
```

#### 記事配信ディレクトリを変更している場合の設定

はてなブログの設定で記事配信ディレクトリを変更している場合は、以下のように設定します。

```yaml
example.hatenablog.com:
  username: sample
  password: <API KEY>
  entry_directory: articles
```

### エントリをダウンロードする（blogsync pull）

設定が完了したら、以下のコマンドを実行すると当該のブログに投稿しているエントリがその URL ローカルに保存されます。固定ページ機能を利用している場合、それもまとめてダウンロードされます。

```sh
% blogsync pull <blogID>
```

この際保存されるファイルのパスは、エントリの URL ベースにしたものとなります。blogsync pull motemen.hatenablog.com した結果だとこんな感じになります（分かりやすいように少し省略しています）:

```
/Users/motemen/Dropbox/Blog/motemen.hatenablog.com/
└── entry
    ├── 2014
    │   ├── 05
    │   │   ├── 12
    │   │   │   └── gulp,_TypeScript,_Browserify_で_Chrome_拡張を書く.md
    │   │   └── 14
    │   │       └── datetime-sh.md
    │   ├── 06
    │   │   ├── 01
    │   │   │   └── introducing-ghq.md
    │   │   ├── 03
    │   │   │   └── git-hub-sync-repo-info.md
…
```

以降は blogsync pull すると、ブログエントリとローカルのファイルをつき合わせ、新しいエントリのみダウンロードされるようになります。

ちなみに、blogIDは省略可能で、省略した場合 `blogsync.yaml` に設定されているブログの内容がpullされます。

### ファイルのフォーマット

エントリのファイルはYAML Frontmatter形式のメタデータではじまり、そののち本文が続く、というフォーマットです:

```
---
Title:   まだmechanizeで消耗してるの? WebDriverで銀行をスクレイピング（ProtractorとWebdriverIOを例に）
Category:
- scraping
Date:    2014-10-01T08:30:00+09:00
URL:     http://motemen.hatenablog.com/entry/2014/10/01/scrape-by-protractor-webdriverio
EditURL: https://blog.hatena.ne.jp/motemen/motemen.hatenablog.com/atom/entry/8454420450066634133
---

今日はスクレイピングの話をします。…
```

今のところメタデータの内容は以下の6つ。

- Title: エントリのタイトル。
- Date: ブログに表示されるエントリの投稿日時。2006-01-02T15:04:05-07:00 といったフォーマットを期待しています。
- URL: エントリの URL。これは自動的に与えられ、書き換えても効果はありません。
- EditURL: エントリを一意に区別する URL。名前のとおり、AtomPub の編集用の URL です。
- Category: エントリーのカテゴリの配列
- Draft: この値が "yes" のとき、下書きとして扱われます。

### エントリを更新する（blogsync push）

ひとたびエントリをダウンロードしたら、そのファイルを編集することで記事を更新できます。

```sh
% blogsync push <path/to/file>
```

例えばこんな感じですね:

```console
% blogsync push ~/Dropbox/blog/motemen.hatenablog.com/entry/2014/12/22/blogsync.md
       GET ---> https://blog.hatena.ne.jp/motemen/motemen.hatenablog.com/atom/entry/8454420450077731341
       200 <--- https://blog.hatena.ne.jp/motemen/motemen.hatenablog.com/atom/entry/8454420450077731341
       PUT ---> https://blog.hatena.ne.jp/motemen/motemen.hatenablog.com/atom/entry/8454420450077731341
       200 <--- https://blog.hatena.ne.jp/motemen/motemen.hatenablog.com/atom/entry/8454420450077731341
     store /Users/motemen/Dropbox/blog/motemen.hatenablog.com/entry/2014/12/22/blogsync.md
```

ファイルがリモートの記事よりも新しくない場合は、更新リクエストは行われません。

基本的にはダウンロードしてきたファイルを更新する用途のコマンドですが、新しくファイルを配置してpushすることも可能です。

#### ファイルパスとURLの関係

エントリーのファイルパスと公開URLのパスは対応しており、push時にエントリーのファイルパスがそのまま公開URLのパスとして使われます。ファイルをリネーム・移動してからpushすると、公開URLもそれに追随して変更されます。

ファイルパスは常にフロントマッターの `CustomPath` より優先されます。食い違う場合はファイルパスが使われ、フロントマッターの `CustomPath` はクリアされます。

### エントリを投稿する（blogsync post）

まだはてなブログ側に存在しない記事を投稿する場合は、投稿用のコマンドで記事を投稿します。

```sh
% blogsync post <blog> < <path/to/file>
```

blogsync post は標準入力からエントリの内容を受けとって投稿し、投稿されたエントリに対応するファイルをダウンロードします。その後は新しく作成されたファイルを編集し、push することでエントリの編集を続けられます。

このコマンドでは `--title=<TITLE>`、`--draft` という引数によって記事タイトルや下書き状態の指定を行えるのでこんな風に雑に、ターミナルから書き始めることもできます…

```console
% blogsync post --draft --title=blogsync motemen.hatenablog.com
さてかきはじめるか…
^D
```

#### 下書きエントリーの扱い

`--draft` オプションを付けて下書きエントリーを投稿した場合、カスタムパス (`--custom-path`) 指定によって保存先が変わります。

- 指定なし: `entry/_draft/{entryID}.md` に保存されます
    - はてなブログは下書きに対して仮のURLを返しますが、これはリクエストのたびに変わるため、blogsyncでは `_draft/` ディレクトリ配下に entryID ベースで保存し、安定したファイル名を保ちます
    - フロントマッターの `URL` フィールドは空になります（仮のURLは記録しません）
    - `_draft/` 内でファイル名を変更してpushしても、元のファイル名に戻されます
- 指定あり: カスタムパスに対応する位置（例: `entry/my-custom-slug.md`）に保存されます
    - `_draft/` の外にあるファイルは、パスの形式に関わらずカスタムパスが設定されていると見なされます
    - ファイルの位置は push しても変わりません

下書きを公開するには以下のいずれかの方法があります:

- `blogsync push --publish <path/to/file>` で `--publish` フラグを付ける
- フロントマッターから `Draft: true` を削除してから `blogsync push <path/to/file>` する

カスタムパスを指定していない場合、公開時にURLが確定し、ファイルは確定したURLに対応する位置に移動されます。

### 特定のエントリを更新する (blogsync fetch)

エントリファイルを指定して、リモートの更新を取り込むことができます。

```sh
% blogsync fetch <path/to/file>
```


### GitHub Actions

`uses: x-motemen/blogsync@v0` とすればblogsyncをインストールできます。

## Author

[motemen](https://github.com/motemen), [Songmu](https://github.com/Songmu)
