# Huh 入門：作りたい質問から調べる

Huh は、Go でターミナル上の質問フォームを作るライブラリです。
ホスト名の入力、言語の選択、パッケージの複数選択などを、画面描画を自作せずに実装できます。
内部では Bubble Tea を使いますが、このガイドでは `Run()` でフォームを実行する方法を扱います。

このリポジトリの **`charm.land/huh/v2 v2.0.3`** を対象にしています。
メソッドとコード例は、この版のモジュールソースに合わせています。
[公式 README](https://github.com/charmbracelet/huh/tree/v2.0.3) と [API リファレンス](https://pkg.go.dev/charm.land/huh/v2@v2.0.3) も参照できます。
実装を進める順序は [実装課題](implementation-exercises.md) を使い、このページは「どう書くか」を調べる資料として使ってください。

## 作りたいものから探す

| やりたいこと | 読む場所 | 主な API |
| --- | --- | --- |
| まず動くフォームを見たい | [動かせるサンプル](#動かせるサンプル) | `NewForm` / `Run` |
| 書き方の意味を知りたい | [基本の構造](#基本の構造) | Field / Group / Form / `Value` |
| ホスト名やパスを一行入力したい | [文字列の入力](#input) | `NewInput` |
| 数字を入力してもらいたい | [数値の入力](#number) | `Input` / `strconv.Atoi` |
| コマンドや説明を複数行で入力したい | [複数行の入力](#text) | `NewText` |
| 言語や方式から一つ選びたい | [単一選択](#select) | `NewSelect[T]` |
| タスクを複数選びたい | [複数選択](#multiselect) | `NewMultiSelect[T]` |
| はい／いいえを質問したい | [真偽値の質問](#confirm) | `NewConfirm` |
| パスワードを隠して入力したい | [パスワード入力](#パスワード入力) | `EchoMode` |
| パスワードを二回入力してもらいたい | [確認入力](#password-confirmation) | 別変数＋`Validate` |
| 説明だけの画面を挟みたい | [説明の表示](#note) | `NewNote` |
| 実行 PC のファイルを選びたい | [ファイル選択](#filepicker) | `NewFilePicker` |
| 空欄や不正な入力を止めたい | [入力検証](#入力検証) | `Validate` |
| 初期値や入力例を表示したい | [初期値と入力例](#defaults) | 変数の初期値 / `Placeholder` |
| カテゴリごとにページを分けたい | [ページ分け](#pages) | `NewGroup` |
| 固定 IP の時だけ追加で質問したい | [条件付き質問](#回答によって質問を変える) | Go の `if` / `WithHideFunc` |
| 国を変えたら候補も変えたい | [候補の連動](#dynamic-options) | `OptionsFunc` |
| 回答を構造体や map に入れたい | [回答の保存先](#answer-storage) | `Value` / `Key` |
| 確認画面から修正できるようにしたい | [確認と修正](#review) | フォームを再構築するループ |
| YAML から質問を作りたい | [YAML との組み合わせ](#yaml-との組み合わせ) | `[]huh.Field` / `switch` |
| テーマや表示サイズを変えたい | [見た目と実行環境](#見た目と実行環境) | `WithTheme` / `WithWidth` |
| キー操作やよくある問題を確認したい | [基本操作](#基本操作) / [困った時](#troubleshooting) | キーマップ / エラー処理 |

## 導入

このリポジトリには依存関係が追加済みです。Go ファイルで次を import します。

```go
import "charm.land/huh/v2"
```

別のディレクトリで練習する場合は、次のように用意します。

```sh
mkdir huh-demo
cd huh-demo
go mod init example.com/huh-demo
go get charm.land/huh/v2@v2.0.3
```

コードを `main.go` に保存し、`go run .` で起動します。
リポジトリで試す場合も、複数ファイルを一緒にコンパイルするため `go run .` を使います。
`go run main.go` は、隣にある別の Go ファイルを自動では含めません。

v1 の記事にある `github.com/charmbracelet/huh` と、このプロジェクトの v2 の import を混在させないでください。
API を調べる時は、例えば `go doc charm.land/huh/v2.Input.Placeholder` とすると手元の版を確認できます。

### このページのコード例の試し方

「動かせるサンプル」は `package main` と import を含む、そのまま実行できるプログラムです。
それ以外は、原則として **`main.go` に追加する関数の例**です。それぞれの直前に必要な import を示します。

1. 関数を `main` の外へ追加する。
2. 必要なパッケージを、既存の import に追加する。
3. `main` の中からその関数を呼ぶ。
4. 戻り値のエラーを確認し、成功した時だけ回答を使う。

例えば後述の `askHostname()` は、次のように呼び出せます。
必要な import は `fmt`、`os` です。`askHostname` 本体は [一行入力の節](#input) から追加します。

```go
func main() {
	hostname, err := askHostname()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println("ホスト名:", hostname)
}
```

この呼出し例の `main` と、別のサンプルの `main` を同じパッケージに二つ置かないでください。
本番アプリの中断と終了コードは、[エラー処理](#errors) で整理します。

## 基本の構造

### Field・Group・Form は何が違うか

| 要素 | 役割 | 作成方法 | 考え方 |
| --- | --- | --- | --- |
| Field | 一つの質問や説明 | `huh.NewInput()` など | 「何を入力してもらうか」 |
| Group | 質問をまとめたページ | `huh.NewGroup(fields...)` | 「どの質問を一緒に見せるか」 |
| Form | ページをまとめたフォーム | `huh.NewForm(groups...)` | 「どの順番で回答してもらうか」 |

作る順序は「回答の型を決める → Field を作る → Group にまとめる → Form を実行する」です。
ホスト名なら `string`、ミラーを使うかなら `bool`、タスクなら `[]string` と、画面より先に回答の型を決めると整理できます。

`Run()` は回答の完了、ユーザーの中断、またはエラーまで待ち、`error` を返します。
Field 一つでも `Run()` できます。質問が複数なら Form にまとめます。

### メソッドをつなげて書く理由

必要な import：`charm.land/huh/v2`。

```go
func makeHostnameField(hostname *string) *huh.Input {
	return huh.NewInput().
		Title("ホスト名").
		Description("インストール先のマシン名です").
		Value(hostname)
}
```

`NewInput()` で入力欄を作り、`Title` で質問名、`Description` で補足、`Value` で回答先を設定しています。
それぞれが設定後の Field を返すため、続けて次のメソッドを呼べます。これをメソッドチェーンと呼びます。

この関数は Field を作るだけです。まだ表示や入力は行いません。
`Run()` は Field または Form のどちらかで呼びます。同じ Field を個別実行してから、同じ質問として Form でも実行する必要はありません。

改行してつなげる時は、上の例のように **行末へ `.` を置きます**。
Go は改行によって文の終わりを判断するため、次の行頭だけに `.` を置く書き方ではエラーになります。

### Value とポインターを理解する

`Value(&hostname)` は、Huh に回答を書き込む変数を渡す指定です。
`&hostname` は変数の場所を渡すポインターで、Huh がその変数を更新できます。
`makeHostnameField` の引数は既に `*string` なので、関数内では `Value(hostname)` と渡しています。

| Field | 用意する変数 | `Value` に渡す型 |
| --- | --- | --- |
| Input / Text / FilePicker | `string` | `*string` |
| Select[string] | `string` | `*string` |
| Select[int] | `int` | `*int` |
| MultiSelect[string] | `[]string` | `*[]string` |
| Confirm | `bool` | `*bool` |
| Note | 回答変数なし | `Value` は使わない |

**重要：** `Value` を渡した変数は、操作途中にも更新され得ます。
`Run()` が成功した時だけ完成した回答として扱います。中断した回答を、そのまま cfg の生成へ渡してはいけません。

### よく使う共通メソッド

| メソッド | 何を決めるか | 注意点 |
| --- | --- | --- |
| `Title("...")` | 質問名 | 何を決める質問か、短く書く |
| `Description("...")` | 補足説明 | 入力例・用途・空欄の意味を伝える |
| `Value(&answer)` | 回答を保存する変数 | Field に合う型のポインターを渡す |
| `Validate(fn)` | 入力が正しいかを判定する関数 | 形式ごとに関数の引数の型が違う |
| `Key("id")` | Form から結果を探すための ID | 変数を渡す `Value` とは別の方法 |
| `Run()` | 質問またはフォームを実行する | エラーを確認してから回答を使う |

すべての Field にすべてのメソッドがあるわけではありません。
例えば `Placeholder` は Input / Text、`Options` は Select / MultiSelect、`Lines` は Text のものです。

## 動かせるサンプル

言語、ホスト名、タスク、保存の確認を二ページで質問します。
ここでの候補は学習用の一部です。Preseed の生成・ファイル保存は含めていません。

```go
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"charm.land/huh/v2"
)

func main() {
	language := "ja"
	hostname := "debian-server"
	tasks := []string{"standard"}
	save := false

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("インストール時の言語").
				Options(
					huh.NewOption("日本語", "ja"),
					huh.NewOption("English", "en"),
				).
				Value(&language),
			huh.NewInput().
				Title("ホスト名").
				Description("この例では空欄だけを検証します").
				Value(&hostname).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("ホスト名を入力してください")
					}
					return nil
				}),
		).Title("基本設定"),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("インストールするタスク").
				Options(
					huh.NewOption("標準ユーティリティ", "standard"),
					huh.NewOption("SSH サーバー", "ssh-server"),
				).
				Value(&tasks),
			huh.NewConfirm().
				Title("この回答を使って保存しますか？").
				Affirmative("保存する").
				Negative("保存しない").
				Value(&save),
		).Title("パッケージと確認"),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Fprintln(os.Stderr, "入力を中断しました")
			return
		}
		fmt.Fprintln(os.Stderr, "フォームの実行に失敗しました:", err)
		os.Exit(1)
	}
	if !save {
		fmt.Println("保存せず終了しました")
		return
	}

	fmt.Printf("言語: %s\nホスト名: %s\nタスク: %v\n", language, hostname, tasks)
	fmt.Println("このサンプルでは、ここまでが回答の確認です")
}
```

まず実行して、初期値を変更する、複数選択を解除する、保存しない、中断する、を試してください。
`Confirm` で「保存する」を選んでも、Huh がファイルを保存するわけではありません。保存処理は自分で呼び出します。

<a id="input"></a>
## 文字列を一行入力したい：Input

### どんな質問に使うか

ホスト名、ドメイン名、IP、ミラーのホスト、パッケージ名の一覧など、**一行の文字列を自由に入力する質問**に使います。
候補が少数で決まっている場合は、Input で自由入力させるより [Select](#select) が適しています。

### メソッド一覧

| メソッド | 使い方 | 考え方 |
| --- | --- | --- |
| `NewInput()` | 入力欄を作る | 回答の型は `string` |
| `Title` / `Description` | 質問と補足を書く | 空欄が何を意味するかも書く |
| `Value(&value)` | 回答先を指定する | 変数の初期値が実際の入力値になる |
| `Placeholder("例: ...")` | 空欄に入力例を表示する | 回答の値にはならない |
| `CharLimit(63)` | 入力の長さを制限する | 許される文字の判定は別途 `Validate` で行う |
| `Suggestions([]string{...})` | 入力の補完候補を設定する | 候補以外の入力もできる |
| `Prompt("> ")` | 入力欄の記号を変える | 質問名を変える `Title` とは別 |
| `Inline(true)` | タイトルと入力欄を横並びにする | 短い質問で使いやすい |
| `Validate(func(string) error)` | 不正な入力を止める | 正常なら `nil` を返す |

詳細：[v2.0.3 の Input 実装](https://github.com/charmbracelet/huh/blob/v2.0.3/field_input.go)。

### ホスト名を入力する例

必要な import：`charm.land/huh/v2`、`errors`、`strings`。

```go
func askHostname() (string, error) {
	hostname := "debian-server"
	err := huh.NewInput().
		Title("ホスト名").
		Description("インストール先のマシン名。空欄にはできません").
		Placeholder("例: debian-server").
		CharLimit(63).
		Value(&hostname).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("ホスト名を入力してください")
			}
			return nil
		}).
		Run()
	if err != nil {
		return "", err
	}
	return hostname, nil
}
```

この例の検証は空欄だけです。実際のホスト名の文字種・先頭末尾・長さの規則は、[実装課題3](implementation-exercises.md#課題3入力検証を関数にする) で追加します。
`CharLimit` は入力操作の補助です。生成前にも同じ制約を確かめます。

### 補完と選択肢を混同しない

ミラーのホストに `Suggestions([]string{"deb.debian.org"})` を付けると、入力の補完候補になります。
既定の操作では `Ctrl+E` で補完できます。別のホスト名も入力できます。
候補に存在する値だけを許可したいなら、`Select` を使うか `Validate` で制限します。

<a id="number"></a>
## 数字を入力したい：Input と数値変換

Input は数値も `string` として受け取ります。
自由入力の数値は、入力中に検証し、成功後に `strconv.Atoi` などで数値へ変換します。
選択肢が「25・50・100」のように固定なら、[Select[int]](#select-number) にすると変換を減らせます。

必要な import：`charm.land/huh/v2`、`errors`、`strconv`。

```go
func askPercentage() (int, error) {
	value := "100"
	err := huh.NewInput().
		Title("LVM に使う割合").
		Description("1〜100の整数。%記号は付けません").
		Value(&value).
		Validate(func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 || n > 100 {
				return errors.New("1〜100の整数を入力してください")
			}
			return nil
		}).
		Run()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}
```

`100%` のような出力文字列へ変えるのは、質問の後の生成処理です。
`CharLimit(3)` だけでは、`abc` や `999` のような不正な値を防げません。

<a id="text"></a>
## コマンドなどを複数行入力したい：Text

### どんな質問に使うか

終了直前の独自コマンド、説明文など、**改行を含む文字列**に使います。
ホスト名や IP に Text を使うと、不要な改行の扱いが増えるため、一行入力と分けます。

### メソッド一覧

| メソッド | 使い方 | 考え方 |
| --- | --- | --- |
| `NewText()` | 複数行入力欄を作る | 回答は改行を含む `string` |
| `Lines(6)` | 表示する入力欄の行数 | 入力できる最大行数ではない |
| `ShowLineNumbers(true)` | 行番号を表示する | 長い入力で位置を把握しやすい |
| `Placeholder("...")` | 入力例を表示する | 例は回答に含まれない |
| `CharLimit(4096)` | 入力文字数を制限する | ファイル出力のバイト数制限とは区別する |
| `Value(&text)` | 初期値と回答先 | 初期値に `\n` を含められる |
| `Validate(func(string) error)` | 本文を検証する | 改行を許すルールを使う |
| `ExternalEditor(true)` | 外部エディターを使えるようにする | 既定キーは `Ctrl+E`。実行 PC 側のエディターが開く |
| `Editor("nano")` | 使用するエディターを指定する | 存在するエディターを選ぶ |
| `EditorExtension("sh")` | 一時編集ファイルの拡張子 | エディターの言語認識に使える |

詳細：[v2.0.3 の Text 実装](https://github.com/charmbracelet/huh/blob/v2.0.3/field_text.go)。

### 終了直前コマンドを入力する例

必要な import：`charm.land/huh/v2`、`errors`、`strings`。

```go
func askLateCommand() (string, error) {
	var command string
	err := huh.NewText().
		Title("終了直前に実行するコマンド").
		Description("空欄なら追加しません。改行は Alt+Enter または Ctrl+J").
		Placeholder("例: in-target touch /root/setup-complete").
		Lines(6).
		ShowLineNumbers(true).
		Value(&command).
		Validate(func(s string) error {
			if strings.ContainsAny(s, "\x00\r") {
				return errors.New("NUL と CR は入力できません")
			}
			return nil
		}).
		Run()
	if err != nil {
		return "", err
	}
	return command, nil
}
```

**通常の `Enter` は次の質問へ進む操作です。** 改行には `Alt+Enter` または `Ctrl+J` を使います。
Text は本文を受け取るだけで、コマンドを実行しません。
Preseed へ出力する時の一行化は生成処理で行い、生成 PC で本文を実行しない設計にします。

<a id="select"></a>
## 候補から一つ選びたい：Select

### どんな質問に使うか

言語、DHCP／固定 IP、パーティション方式、再起動／停止／電源OFFなど、**一つだけ選ぶ質問**に使います。
画面に見せる名前と、回答として保存する値を分けられます。

### メソッド一覧

| メソッド | 使い方 | 考え方 |
| --- | --- | --- |
| `NewSelect[string]()` | 文字列を一つ選ぶ欄 | `[string]` は回答値の型 |
| `Options(option1, option2)` | 候補を指定する | 各候補を `NewOption` で作る |
| `NewOption("表示名", "値")` | 表示名と回答値を組にする | Preseed には値を使う |
| `Value(&value)` | 初期値と回答先 | 初期値は候補内の値にする |
| `NewOption(...).Selected(true)` | 初期選択を候補側で指定する | 変数の初期値と二重に管理しない |
| `Height(8)` | 表示の高さを指定する | 長い候補一覧をスクロールさせる |
| `Filtering(true)` | 最初から絞込み入力の状態にする | `true` は「絞込みを利用可能にする」だけの意味ではない |
| `Inline(true)` | 横並びの選択表示にする | 少数の短い選択肢に向く |
| `Validate(func(string) error)` | 選択値を検証する | 選んだ候補の追加条件を確認する |
| `OptionsFunc(fn, binding)` | 回答に応じて候補を変える | [候補の連動](#dynamic-options) を参照 |

詳細：[v2.0.3 の Select 実装](https://github.com/charmbracelet/huh/blob/v2.0.3/field_select.go)。

### 表示は日本語、回答はコードにする例

必要な import：`charm.land/huh/v2`。

```go
func askLanguage() (string, error) {
	language := "ja"
	err := huh.NewSelect[string]().
		Title("インストール時の言語").
		Options(
			huh.NewOption("日本語", "ja"),
			huh.NewOption("English", "en"),
		).
		Value(&language).
		Run()
	if err != nil {
		return "", err
	}
	return language, nil
}
```

`NewOption("日本語", "ja")` は、画面に「日本語」を表示して、変数に `ja` を保存します。
初期値を `"日本語"` にすると、候補の値とは一致しません。表示名でなく値を使います。
Select は未選択を必須入力のようには表現しません。初期値が候補にないと先頭などへ選択が移り得るため、定義の初期値と候補の整合性はアプリ側で検証します。

<a id="select-number"></a>
### 数値の候補なら Select[int] を使う

必要な import：`charm.land/huh/v2`。

```go
func askPresetPercentage() (int, error) {
	percentage := 100
	err := huh.NewSelect[int]().
		Title("LVM に使う割合").
		Options(
			huh.NewOption("半分（50%）", 50),
			huh.NewOption("全体（100%）", 100),
		).
		Value(&percentage).
		Run()
	if err != nil {
		return 0, err
	}
	return percentage, nil
}
```

`[int]` は「この選択欄の値を int にする」という指定です。Go のジェネリクスを使っています。
候補の値、回答変数、`Validate` の引数も同じ型に揃えます。
実際の YAML から読み込む値が文字列なら、選択欄も `string` に揃えて、必要な場面で変換する方法でも構いません。

### スライスに入った候補を渡す

必要な import：`charm.land/huh/v2`。

```go
func askFromOptions(options []huh.Option[string]) (string, error) {
	var value string
	err := huh.NewSelect[string]().
		Title("一つ選んでください").
		Options(options...).
		Value(&value).
		Run()
	if err != nil {
		return "", err
	}
	return value, nil
}
```

`options...` は、スライスの要素を複数の引数として渡す書き方です。
この例の呼出し側では、**候補が一つ以上あることを確認してから**呼びます。空候補を正常な質問として実行しないでください。

<a id="multiselect"></a>
## 候補から複数選びたい：MultiSelect

### どんな質問に使うか

インストールするタスク、追加 locale、security／updates など、**複数の候補を同時に選べる質問**に使います。
回答は `[]string` などのスライスになります。選択しない状態を許すかどうかも決めます。

### メソッド一覧

| メソッド | 使い方 | 考え方 |
| --- | --- | --- |
| `NewMultiSelect[string]()` | 文字列を複数選ぶ欄 | 回答の型は `[]string` |
| `Options(...)` | 候補を指定する | Select と同じ `NewOption` を使う |
| `Value(&values)` | 初期値と回答先 | 初期値は候補内の値のスライス |
| `NewOption(...).Selected(true)` | 候補を初期チェック済みにする | 変数の初期値と併用する時は矛盾を避ける |
| `Limit(3)` | 選択できる個数の上限 | 最小個数の指定ではない |
| `Height(8)` | 表示の高さ | 候補数が多い場合に使う |
| `Filterable(false)` | 絞込みを使わせない | 通常は既定のままでよい |
| `Filtering(true)` | 絞込み入力から開始する | 利用可否を変える `Filterable` と区別する |
| `Validate(func([]string) error)` | 選択結果全体を検証する | 最低一つ、特定の組合せ禁止など |
| `OptionsFunc(fn, binding)` | 候補を動的に変える | 元の回答がまだ有効かも確認する |

詳細：[v2.0.3 の MultiSelect 実装](https://github.com/charmbracelet/huh/blob/v2.0.3/field_multiselect.go)。

### タスクを複数選ぶ例

必要な import：`charm.land/huh/v2`。

```go
func askTasks() ([]string, error) {
	tasks := []string{"standard"}
	err := huh.NewMultiSelect[string]().
		Title("インストールするタスク").
		Description("Space でチェック。何も選ばなくても構いません").
		Options(
			huh.NewOption("標準ユーティリティ", "standard"),
			huh.NewOption("SSH サーバー", "ssh-server"),
			huh.NewOption("Web サーバー", "web-server"),
		).
		Value(&tasks).
		Height(8).
		Run()
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
```

初期状態で `standard` にチェックが入り、ユーザーはそれを外せます。
候補を移動する操作と、チェックする操作は別です。通常は上下キーで移動し、Space でチェックを切り替えます。

### 最低一つ選んでもらいたい

必要な import：`errors`。上の Field の `Run()` より前に `.Validate(requireOneTask)` を追加します。

```go
func requireOneTask(values []string) error {
	if len(values) == 0 {
		return errors.New("一つ以上選んでください")
	}
	return nil
}
```

これは「最低一つを必須にしたい別の質問」の例です。このプロジェクトのタスクで空選択を許すなら、付けません。
`Limit(1)` は「最大一つ」であり、ゼロ個を拒否しません。一つしか選ばせない目的なら、Select の方が自然です。
出力時の `standard, ssh-server` への連結は、Huh の担当ではありません。生成処理で区切り文字を決めます。

<a id="confirm"></a>
## はい／いいえを質問したい：Confirm

### どんな質問に使うか

ミラーを使うか、NTP を有効にするか、保存するかなど、**答えが真偽値になる質問**に使います。
「再起動・停止・電源OFF」のような三択や、「生成・修正・終了」の操作メニューは Select にします。

### メソッド一覧

| メソッド | 使い方 | 考え方 |
| --- | --- | --- |
| `NewConfirm()` | 真偽値の質問を作る | 回答は `bool` |
| `Value(&value)` | 初期状態と回答先 | `true` / `false` の初期値を持つ |
| `Affirmative("使用する")` | true 側の表示 | true の意味が分かる言葉にする |
| `Negative("使用しない")` | false 側の表示 | false がエラーという意味ではない |
| `Inline(true)` | タイトルと選択を横並びにする | 短い質問に向く |
| `Validate(func(bool) error)` | 回答を検証する | 通常の確認では false も受け付ける |

詳細：[v2.0.3 の Confirm 実装](https://github.com/charmbracelet/huh/blob/v2.0.3/field_confirm.go)。

### ミラー使用を質問する例

必要な import：`charm.land/huh/v2`。

```go
func askUseMirror() (bool, error) {
	useMirror := true
	err := huh.NewConfirm().
		Title("ネットワークミラーを使用しますか？").
		Affirmative("使用する").
		Negative("使用しない").
		Value(&useMirror).
		Run()
	if err != nil {
		return false, err
	}
	return useMirror, nil
}
```

`false, nil` は「使用しないと正常に回答した」、`false, err` は「入力に失敗した」の意味です。
戻り値の bool だけを見ず、必ずエラーも確認します。

保存の確認なら、初期値を `false` にして「保存する／保存しない」と表示します。
「保存しない」を選んだ時に終了するか修正へ戻るかは、呼出し側で `if` を書いて決めます。
`Validate` で false を拒否すると、「保存しない」を選んで終了することもできなくなるため、用途を考えて使います。

## パスワード入力

### 入力文字の表示だけを変える

パスワードも回答の型は `string` なので Input を使います。表示の方法を `EchoMode` で変えます。

| 設定 | 表示 |
| --- | --- |
| `huh.EchoModeNormal` | 文字をそのまま表示する |
| `huh.EchoModePassword` | 文字をマスクして表示する |
| `huh.EchoModeNone` | 入力文字もマスク文字も表示しない |

このリポジトリでは、非表示入力なら `EchoModeNone` を使います。
旧 API の `Password(true)` より、表示の意図が分かる `EchoMode` を使ってください。
詳細：[v2.0.3 の Input 実装](https://github.com/charmbracelet/huh/blob/v2.0.3/field_input.go)。

<a id="password-confirmation"></a>
### パスワードと確認用を二回入力する

必要な import：`charm.land/huh/v2`、`errors`。

```go
func askPassword() (string, error) {
	var password string
	var confirmation string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("ユーザーパスワード").
				EchoMode(huh.EchoModeNone).
				Value(&password).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("パスワードを入力してください")
					}
					return nil
				}),
			huh.NewInput().
				Title("パスワードをもう一度入力").
				EchoMode(huh.EchoModeNone).
				Value(&confirmation).
				Validate(func(s string) error {
					if s != password {
						return errors.New("パスワードが一致しません")
					}
					return nil
				}),
		),
	)
	if err := form.Run(); err != nil {
		return "", err
	}
	if password == "" || confirmation != password {
		return "", errors.New("パスワードの確認をやり直してください")
	}
	return password, nil
}
```

確認入力に同じ変数を使うと、最初の入力値を上書きして比較できなくなります。別の変数を使います。
ここでの `func(s string) error` は、その場で作る関数です。外側の `password` も参照できるため、二つの値を比較できます。
最初の入力へ戻って変更した場合も考え、フォーム終了後にもう一度一致を確認しています。

Huh は、非表示の文字列を受け取るところまで担当します。
ハッシュ化は別処理です。Preseed には予定している SHA-512 crypt を出力し、平文・ハッシュを回答一覧やログへ表示しません。
また、Huh の入力に `x/term.ReadPassword` を重ねて使う必要はありません。

<a id="note"></a>
## 説明だけの画面を挟みたい：Note

導入説明や注意事項など、**回答を必要としない表示**には Note を使います。
同意を回答として残したい場合は Confirm にします。Note に表示しただけでは、同意の bool は得られません。

| メソッド | 使い方 |
| --- | --- |
| `NewNote()` | 説明の Field を作る |
| `Title` / `Description` | 見出しと本文を書く |
| `Next(true)` | 次へ進むボタンを表示する |
| `NextLabel("設定を始める")` | ボタンの表示を変える |
| `Height(8)` | 表示の高さを指定する |

必要な import：`charm.land/huh/v2`。

```go
func showIntroduction() error {
	return huh.NewNote().
		Title("Debian の設定を始めます").
		Description("インストール先の情報を回答してください。このアプリは cfg を生成します").
		Next(true).
		NextLabel("設定を始める").
		Run()
}
```

アクセシブルモードの Note は説明を出力して進むため、同じ待機ボタンの動作を前提にしません。
Note と、Preseed の出力型 `note`、YAML の `answer.kind: none` はそれぞれ別の概念です。
`none` の固定値は、説明画面を作らず生成へ渡せます。
詳細：[v2.0.3 の Note 実装](https://github.com/charmbracelet/huh/blob/v2.0.3/field_note.go)。

<a id="filepicker"></a>
## 実行 PC のファイルを一覧から選びたい：FilePicker

外部 YAML など、**既に実行 PC 上に存在するファイル**を選ぶ時に使えます。
インストール先のディスク名や NIC 名を選ぶためには使いません。
生成する新しいファイル名の自由入力には Input が向いています。

| メソッド | 使い方 |
| --- | --- |
| `NewFilePicker()` | ファイル選択欄を作る |
| `CurrentDirectory(".")` | 一覧を開くディレクトリ |
| `AllowedTypes([]string{".yaml", ".yml", ".YAML"})` | 選択可能な拡張子 |
| `FileAllowed(true)` | ファイル選択を許可する |
| `DirAllowed(false)` | ディレクトリそのものを回答にしない |
| `ShowHidden(true)` | 隠しファイルを表示する |
| `Value(&path)` | 選んだパスを保存する |
| `Validate(func(string) error)` | パスに対する追加検証 |

必要な import：`charm.land/huh/v2`。

```go
func askDefinitionFile() (string, error) {
	var path string
	err := huh.NewFilePicker().
		Title("質問定義の YAML を選択").
		CurrentDirectory(".").
		AllowedTypes([]string{".yaml", ".yml", ".YAML"}).
		FileAllowed(true).
		DirAllowed(false).
		Value(&path).
		Run()
	if err != nil {
		return "", err
	}
	return path, nil
}
```

拡張子の制限は YAML の内容の正しさを保証しません。
選択後にファイルを読み、そのエラーとスキーマの検証を別に行います。
詳細：[v2.0.3 の FilePicker 実装](https://github.com/charmbracelet/huh/blob/v2.0.3/field_filepicker.go)。

## 入力検証

### Validate は「正しければ nil、直してほしければ error」

| Field | 検証関数の型 |
| --- | --- |
| Input / Text / FilePicker | `func(string) error` |
| Select[string] | `func(string) error` |
| Select[int] | `func(int) error` |
| MultiSelect[string] | `func([]string) error` |
| Confirm | `func(bool) error` |

Huh は検証のエラーを表示し、不正な値のまま次へ進むのを止めます。
「IPv4 として正しい」「ユーザー名として許される」などのルールは、自分で用意します。
検証関数にファイル保存やコマンド実行などの副作用を入れず、値の判定だけを担当させます。

### 関数を「その場で呼ぶ」と「後で呼べるように渡す」の違い

`.Validate(validateIPv4)` は、関数そのものを Huh に渡します。Huh が入力値を引数にして、その関数を呼びます。
ここに `validateIPv4(address)` と書くと、その場で検証した結果の `error` を渡すことになり、要求される型が違います。

一方、`.Validate(huh.ValidateNotEmpty())` の `()` は必要です。
`ValidateNotEmpty()` は「検証結果」ではなく「検証する関数」を作って返す関数なので、呼び出してから渡します。
その場で書く `func(s string) error { ... }` も、検証関数を作って渡す書き方です。

### 同じ検証を入力欄と生成前で使う

必要な import：`charm.land/huh/v2`、`errors`、`net`。

```go
func validateIPv4(s string) error {
	ip := net.ParseIP(s)
	if ip == nil || ip.To4() == nil {
		return errors.New("IPv4アドレスを入力してください")
	}
	return nil
}

func askIPv4() (string, error) {
	var address string
	err := huh.NewInput().
		Title("IPv4アドレス").
		Placeholder("例: 192.168.10.20").
		Value(&address).
		Validate(validateIPv4).
		Run()
	if err != nil {
		return "", err
	}
	if err := validateIPv4(address); err != nil {
		return "", err
	}
	return address, nil
}
```

この関数は書式だけの例です。IPv4-mapped IPv6 表記の厳密な制限、ユニキャストの範囲、ネットマスクやゲートウェイとの整合性は別途追加します。
複数の回答の組合せは、フォーム終了後にも検証します。
入力欄の検証を通った後で、親の回答や候補が変更されることがあるためです。

### 組込みの検証関数も使える

| メソッドではなく関数 | 用途 |
| --- | --- |
| `huh.ValidateNotEmpty()` | 空文字列を拒否する |
| `huh.ValidateMinLength(1)` | 最小文字数 |
| `huh.ValidateMaxLength(63)` | 最大文字数 |
| `huh.ValidateLength(1, 63)` | 文字数の範囲 |
| `huh.ValidateOneOf("ja", "en")` | 指定した文字列だけを許可する |

例えば `.Validate(huh.ValidateNotEmpty())` と渡します。
`ValidateNotEmpty` は空白だけの文字列までは拒否しません。空白も拒否したい場合は `strings.TrimSpace` などを使って自分で判定します。
組込み関数のメッセージは英語なので、日本語の理由を示したい時も自作関数を使えます。
詳細：[v2.0.3 の検証関数](https://github.com/charmbracelet/huh/blob/v2.0.3/validate.go)。

<a id="defaults"></a>
## 初期値・入力例・固定値を使い分けたい

| 欲しい動作 | 実装 | 回答に入るか |
| --- | --- | --- |
| 編集できる初期ホスト名 | 変数を `"debian-server"` で初期化して `Value` | 入る。ユーザーが変更できる |
| 空欄に「例: debian-server」を表示 | `Placeholder` | 入らない |
| 初期状態で「日本語」を選ぶ | 回答変数を候補の値 `"ja"` にする | 入る |
| タスクを初期チェックする | `[]string{"standard"}` を渡す | 入る。チェック解除できる |
| 候補側で初期チェックを付ける | `NewOption(...).Selected(true)` | 選択結果になる |
| ミラー使用を初期で有効にする | `useMirror := true` を渡す | 入る |
| ユーザーに聞かず UTC 時計を固定 | Huh の Field を作らず、生成側で `true` を渡す | アプリ側の固定値になる |

**初期値は回答、Placeholder は表示だけ、固定値は質問しない設定です。**
例えば `Placeholder("auto")` にしても、空欄で進めば回答は `auto` になりません。
空欄を `auto` と解釈したいなら、変数を `auto` で初期化するか、回答後の変換規則を明示します。

初期値の指定場所は、なるべく回答変数側に揃えると、YAML の `default` と対応させやすくなります。
変数の初期値と `Selected(true)` を競合させたり、複数選択で以前のチェックを使い回したりしないようにします。

<a id="pages"></a>
## カテゴリごとにページを分けたい：Group

一つの Group に、同じ目的の質問をまとめます。標準レイアウトは一度に一つの Group を表示します。
全67設定を一つの Group へ入れるより、言語・ネットワーク・アカウントなどに分けると見通しがよくなります。

必要な import：`charm.land/huh/v2`。

```go
func askBasicPages() (string, string, error) {
	hostname := "debian-server"
	timezone := "Asia/Tokyo"
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("ホスト名").Value(&hostname),
		).Title("ネットワーク").Description("インストール先の名前を設定します"),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("タイムゾーン").
				Options(
					huh.NewOption("日本", "Asia/Tokyo"),
					huh.NewOption("UTC", "UTC"),
				).
				Value(&timezone),
		).Title("時計"),
	)
	if err := form.Run(); err != nil {
		return "", "", err
	}
	return hostname, timezone, nil
}
```

Group の `Title` はページ名、Field の `Title` は質問名です。
`Tab` / `Shift+Tab` は Form 内の移動で使えますが、別々に `Run()` した Form を自動で行き来する機能ではありません。

一問ずつ別の Form として実行する構成は、条件を順番に評価しやすい利点があります。
一つの Form にまとめる構成は、フォーム内の前後移動を使いやすい利点があります。
最初は処理しやすい方を選び、必要になった段階でカテゴリの再編集を追加します。

## 回答によって質問を変える

### 最初は Go の if でフォームを分ける

固定 IPv4 を使うか先に質問し、必要なら追加のフォームを実行します。
質問の作成前に回答が確定するため、初学者には追いやすい方法です。
必要な import：`charm.land/huh/v2`。

```go
func askNetworkInStages() (bool, string, error) {
	var useStatic bool
	if err := huh.NewConfirm().
		Title("固定IPv4を使いますか？").
		Value(&useStatic).
		Run(); err != nil {
		return false, "", err
	}

	if !useStatic {
		return false, "", nil
	}
	var address string
	if err := huh.NewInput().
		Title("IPv4アドレス").
		Value(&address).
		Run(); err != nil {
		return false, "", err
	}
	return true, address, nil
}
```

この例は分岐の書き方だけを示しています。実装ではアドレスの検証と、マスク・ゲートウェイ等の入力を加えます。
別フォームなので、二問目で `Shift+Tab` を押しても一問目のフォームへは戻りません。戻る操作が必要なら、自分でフローを作ります。

### 一つの Form でページを省く：WithHideFunc

必要な import：`charm.land/huh/v2`。

```go
func askNetworkWithHiddenGroup() (bool, string, error) {
	var useStatic bool
	var address string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("固定IPv4を使いますか？").
				Value(&useStatic),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("IPv4アドレス").
				Value(&address),
		).WithHideFunc(func() bool {
			return !useStatic
		}),
	)
	if err := form.Run(); err != nil {
		return false, "", err
	}
	if !useStatic {
		address = ""
	}
	return useStatic, address, nil
}
```

`WithHideFunc` の戻り値が **true なら隠す**ため、`return !useStatic` としています。
`.WithHide(!useStatic)` は作成時点の bool を渡すだけで、その後の回答変更への追従を目的に使うものではありません。
後で変わる値を使う時は関数を渡します。
詳細：[v2.0.3 の Group 実装](https://github.com/charmbracelet/huh/blob/v2.0.3/group.go)。

### 隠した質問の古い回答も処理する

質問を隠すことと、回答を削除することは別です。
固定 IP を入力した後で DHCP へ変更しても、変数に前の IP が残る場合があります。
上の例は完了後に `address` を空にしています。生成処理も `useStatic` を条件にして、不要な行を出さないようにします。

実際のアプリでは「無効な回答を削除する」「依存する候補を再評価する」「必要になった項目を再質問する」を共通処理にします。
パスワード文字の非表示は `EchoMode`、質問ページの非表示は `WithHideFunc` と、用途も分けてください。

<a id="dynamic-options"></a>
## 国や言語の回答に応じて候補を変えたい：OptionsFunc

### 固定の候補と動的な候補

`Options(...)` は、作成時に用意した候補を使います。
`OptionsFunc(fn, binding)` は、参照する値の変化に応じて候補を計算し直します。
まず候補を計算する普通の Go 関数を作り、その後で Huh に渡すと整理できます。

必要な import：`charm.land/huh/v2`。

```go
func timezoneOptions(country string) []huh.Option[string] {
	if country == "JP" {
		return []huh.Option[string]{
			huh.NewOption("日本", "Asia/Tokyo"),
			huh.NewOption("UTC", "UTC"),
		}
	}
	return []huh.Option[string]{huh.NewOption("UTC", "UTC")}
}

func askCountryAndTimezone() (string, string, error) {
	country := "JP"
	timezone := "UTC"
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("国").
				Options(
					huh.NewOption("日本", "JP"),
					huh.NewOption("その他（この例では UTC のみ）", "OTHER"),
				).
				Value(&country),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("タイムゾーン").
				OptionsFunc(func() []huh.Option[string] {
					return timezoneOptions(country)
				}, &country).
				Value(&timezone),
		),
	)
	if err := form.Run(); err != nil {
		return "", "", err
	}
	return country, timezone, nil
}
```

この例の `OTHER` は説明用の値で、実際の国コードではありません。
全候補を扱うアプリでは、YAML の catalog を使います。

### 第二引数の binding は何か

第一引数の関数が「候補の作り方」、第二引数の `&country` が「何が変わったら作り直すか」です。
`country` の値そのものではなくポインターを渡し、後の変化も追えるようにします。
複数の回答に依存する場合は、その複数を反映する binding を用意するか、段階的なフォームで回答確定後に作り直す方法にします。

候補を返す関数には、ファイル書込みや回答の変更などの副作用を入れません。
この版では結果がキャッシュされることもあるため、「常に呼ばれる」ことを前提にした処理を書かないでください。
詳細：[v2.0.3 の Select 実装](https://github.com/charmbracelet/huh/blob/v2.0.3/field_select.go) と [動的評価の実装](https://github.com/charmbracelet/huh/blob/v2.0.3/eval.go)。

### 質問名や説明も回答に合わせて変えられる

Input・Text・Select などには `TitleFunc(fn, binding)` と `DescriptionFunc(fn, binding)` があります。
例えば `DescriptionFunc` で、選んだ国に応じた説明を返せます。
候補の更新と同様に、表示文字列を返すだけの関数を使います。

候補が変わって古い選択値が使えなくなる場合は、候補内にあるか再検証します。
Huh の見た目が更新されたことだけを、回答の整合性の保証にはしません。

<a id="answer-storage"></a>
## 回答を構造体・map・ID に保存したい

### 構造体のフィールドへ直接保存する

必要な import：`charm.land/huh/v2`。

```go
type BasicAnswers struct {
	Hostname  string
	UseMirror bool
}

func askIntoStruct() (BasicAnswers, error) {
	answers := BasicAnswers{Hostname: "debian-server", UseMirror: true}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("ホスト名").Value(&answers.Hostname),
			huh.NewConfirm().Title("ミラーを使う").Value(&answers.UseMirror),
		),
	)
	if err := form.Run(); err != nil {
		return BasicAnswers{}, err
	}
	return answers, nil
}
```

固定の質問が少数なら、構造体は回答の型と役割が分かりやすくなります。
質問数や ID を YAML で増やせるようにする段階で、map を使う回答ストアへ移行できます。

### map の値はローカル変数を経由する

必要な import：`charm.land/huh/v2`。

```go
func askHostnameIntoMap(answers map[string]string) error {
	value := answers["hostname"]
	if err := huh.NewInput().
		Title("ホスト名").
		Value(&value).
		Run(); err != nil {
		return err
	}
	answers["hostname"] = value
	return nil
}
```

呼出し側で `make(map[string]string)` などにより map を用意しておきます。
map の要素には `&answers["hostname"]` のようにポインターを取れないため、ローカル変数で受けてから書き戻します。
これなら、中断した質問の結果を map へ確定せずに済みます。

真偽値や複数選択も扱うアプリでは、すべてを文字列にせず、回答の型を区別する設計にします。
map の `value, ok := answers[id]` の `ok` により、未回答と空の回答を区別できます。

### Key で Form の結果から取得する方法

必要な import：`charm.land/huh/v2`。

```go
func askWithKey() (string, error) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("ホスト名").Key("hostname"),
		),
	)
	if err := form.Run(); err != nil {
		return "", err
	}
	return form.GetString("hostname"), nil
}
```

`Key` は質問の ID、`GetString` はその ID の文字列結果を取得する方法です。
`GetBool`、`GetInt`、汎用の `Get` もあります。
この版の `GetString` / `GetBool` は、ID 不在や型違いでもゼロ値を返すため、必須回答の存在確認の代用にはしません。

このプロジェクトでは、まず `Value` を使って型を意識する方法がおすすめです。
`Key` は回答変数のポインターではなく、条件参照や YAML の setting ID を自動で解釈する機能でもありません。

<a id="review"></a>
## 回答を確認して、修正へ戻りたい

確認画面には、秘密値を隠した回答一覧と「生成／修正／終了」の選択を用意します。
Huh は操作の選択を受け取りますが、修正へ戻る流れは Go のループで実装します。

必要な import：`charm.land/huh/v2`、`fmt`。

```go
func askAndReviewHostname() (string, bool, error) {
	hostname := "debian-server"
	for {
		draft := hostname
		if err := huh.NewInput().
			Title("ホスト名").
			Value(&draft).
			Run(); err != nil {
			return "", false, err
		}
		hostname = draft

		var action string
		if err := huh.NewSelect[string]().
			Title("回答を確認してください").
			Description(fmt.Sprintf("ホスト名: %s", hostname)).
			Options(
				huh.NewOption("この内容で生成する", "generate"),
				huh.NewOption("ホスト名を修正する", "edit"),
				huh.NewOption("終了する", "quit"),
			).
			Value(&action).
			Run(); err != nil {
			return "", false, err
		}
		switch action {
		case "generate":
			return hostname, true, nil
		case "quit":
			return "", false, nil
		}
	}
}
```

`edit` ならループして、現在の回答を初期値にした新しい Field を作ります。
戻り値の bool は「生成へ進むか」であり、この関数では cfg を生成しません。

本番のカテゴリ修正では、回答のコピーを編集し、中断時はそのコピーを破棄します。
親の回答が変わった時の依存項目の再検証は、[実装課題20](implementation-exercises.md#課題20確認画面から回答を修正できるようにする) を参照してください。
`Confirm` の「いいえ」や `Shift+Tab` が、アプリ全体の編集フローを自動で作るわけではありません。

## YAML との組み合わせ

### YAML の answer.kind を Field へ対応させる

Huh に YAML を渡すだけで、自動で質問が完成するわけではありません。
YAML を Go の構造体へ読み、質問の形式に応じて Field を作る処理を、このアプリで用意します。

| YAML の値 | 作る Field | 回答の型 | このアプリで追加する処理 |
| --- | --- | --- | --- |
| `input` | `NewInput` | `string` | 検証と出力変換 |
| `text` | `NewText` | `string` | 改行を含む値の変換 |
| `select` | `NewSelect[string]` | `string` | 候補・初期値の解決 |
| `multiselect` | `NewMultiSelect[string]` | `[]string` | 初期選択・出力の連結 |
| `boolean` | `NewConfirm` | `bool` | true / false への直列化 |
| `none` | 作らない | 固定値や導出値の型 | 画面を出さず値を解決 |

`hidden` は Input の表示方法、`confirm` は確認用 Input を追加する指定として実装します。
`when` は質問と出力の条件、`choices_from` は候補の参照としてアプリ側で解釈します。
既存の `BasePreseed.YAML` は設計案であり、これらの実装はこれから行います。

### Field を配列へ追加して Group にする

次の例は Input だけを動的に作る最小例です。YAML 読込み自体は含めていません。
必要な import：`charm.land/huh/v2`。

```go
type InputQuestion struct {
	ID    string
	Label string
}

func askInputQuestions(questions []InputQuestion) (map[string]string, error) {
	if len(questions) == 0 {
		return map[string]string{}, nil
	}
	values := make([]string, len(questions))
	fields := make([]huh.Field, 0, len(questions))
	for i, question := range questions {
		fields = append(fields,
			huh.NewInput().
				Title(question.Label).
				Value(&values[i]),
		)
	}
	form := huh.NewForm(huh.NewGroup(fields...))
	if err := form.Run(); err != nil {
		return nil, err
	}
	answers := make(map[string]string, len(questions))
	for i, question := range questions {
		answers[question.ID] = values[i]
	}
	return answers, nil
}
```

呼出し側では質問 ID の重複を拒否します。
`[]huh.Field` には、Input や Select など異なる質問形式をまとめられます。
`fields...` はスライスを複数の引数に展開し、`NewGroup` へ渡します。

回答先の `values` は質問数で先に確保しています。
同じ一つの変数へ全 Field を結び付けると、別々の回答を保持できません。
この例を足場に `kind` ごとの `switch` を追加し、複数の型を保持する回答ストアへ進めます。

### Huh と生成処理の境界

```text
YAML を読む
  → 定義・参照・初期値を検証する
  → Huh の質問を作る
  → 回答を検証する
  → 条件と導出値を再評価する
  → 秘密値を隠して確認する
  → Preseed の owner・key・type・value を組み立てる
  → text/template で cfg を生成する
```

例えば、`[]string{"standard", "ssh-server"}` を `standard, ssh-server` へ変えるのは生成処理です。
画面を作るコードに Preseed の行を直接書き込む処理を混ぜないと、画面を開かずに変換をテストできます。

## 見た目と実行環境

### 見た目より先に質問の意味を伝える

最初は標準テーマで構いません。
`Title`、`Description`、Group の分類、初期値、検証エラーが揃うと、操作の目的が分かりやすくなります。
候補が多い時は Select / MultiSelect の `Height` を使い、入力欄の大きさは Text の `Lines` で調整します。

### Form の表示設定

| メソッド | 用途 |
| --- | --- |
| `WithWidth(80)` | フォームの表示幅を指定する |
| `WithHeight(24)` | フォームの高さを指定する |
| `WithShowHelp(true)` | キー操作の案内を表示する |
| `WithShowErrors(true)` | 検証エラーを表示する |
| `WithLayout(huh.LayoutDefault)` | 一度に一つの Group を表示する標準レイアウト |
| `WithLayout(huh.LayoutStack)` | Group を縦に並べるレイアウト |
| `WithLayout(huh.LayoutColumns(2))` | Group を二列に並べるレイアウト |
| `WithTheme(...)` | テーマを指定する |
| `WithAccessible(true)` | 画面再描画を使わない対話モードにする |

列数を増やすと狭い端末で読みづらくなるため、初期実装は標準のページ表示にします。

### v2.0.3 に合わせてテーマを指定する

この版の `WithTheme` は `huh.Theme` を受け取ります。
組込みのテーマ関数を `huh.ThemeFunc` で包んで渡します。
必要な import：`charm.land/huh/v2`。

```go
func askWithTheme() (string, error) {
	var hostname string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("ホスト名").Value(&hostname),
		),
	).
		WithTheme(huh.ThemeFunc(huh.ThemeDracula)).
		WithShowHelp(true).
		WithShowErrors(true)
	if err := form.Run(); err != nil {
		return "", err
	}
	return hostname, nil
}
```

`huh.ThemeCharm`、`huh.ThemeDracula`、`huh.ThemeCatppuccin` などを使えます。
`ThemeFunc` は明暗の情報を受け取り、スタイルを返す関数を Theme として扱うための型です。
この版では `WithTheme(huh.ThemeDracula())` や `WithTheme(huh.ThemeDracula(true))` をそのまま渡す形にはしません。
記事や移行説明だけで判断せず、[v2.0.3 の theme.go](https://github.com/charmbracelet/huh/blob/v2.0.3/theme.go) と手元のコンパイラーで確認してください。

### アクセシブルモードにする

必要な import：`charm.land/huh/v2`。

```go
func askAccessibly() (string, error) {
	var hostname string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("ホスト名").Value(&hostname),
		),
	).WithAccessible(true)
	if err := form.Run(); err != nil {
		return "", err
	}
	return hostname, nil
}
```

設定する場所は Field ではなく Form です。
スクリーンリーダーなどで使いやすい対話方式ですが、無人実行モードではありません。
通常の TUI と表示・キー操作が同じとは限りません。両方で質問・中断・秘密値の表示を確認します。
特に v2.0.3 のアクセシブル実行は、通常の TUI と異なる経路で Field を順番に実行します。
`WithHideFunc` のスキップや `OptionsFunc` の更新が同じように働くことを前提にせず、条件で必要な質問を選び、候補を確定してから Form を作る方法を使います。
このモードでも、生成前にアプリ側で回答の存在・型・整合性を検証します。
詳細：[v2.0.3 の Form 実装](https://github.com/charmbracelet/huh/blob/v2.0.3/form.go)。

<a id="errors"></a>
### 中断・通常の失敗・正常な回答を区別する

必要な import：`errors`、`fmt`、`os`、`charm.land/huh/v2`。

```go
func runWithExitCode(form *huh.Form) int {
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Fprintln(os.Stderr, "入力を中断しました")
			return 130
		}
		fmt.Fprintln(os.Stderr, "フォームを実行できませんでした:", err)
		return 1
	}
	return 0
}
```

呼出し側は、戻り値が0の時だけ確認・生成へ進みます。
最終的なアプリでは `run() error` などでエラーを上へ返し、`os.Exit` は `main` の最後で行います。
`os.Exit` は `defer` を実行しないため、深い関数で呼ぶと片付けを省いてしまいます。

フォームの実行中に `fmt.Println` で表示すると、画面描画と混ざります。
結果は `Run()` の正常終了後に表示します。調査用ログが必要でも、秘密値を含む構造体全体は出力しません。

## 基本操作

通常の TUI の標準キーマップの主な操作です。画面に表示される案内も確認してください。

| キー | 操作 |
| --- | --- |
| `↑` / `↓` | 選択リスト内の移動 |
| `Enter` | 次の質問へ進む、または最後の質問を送信する |
| `Tab` / `Shift+Tab` | 次／前の質問へ移動する（移動可能な場合） |
| `Space` / `x` | MultiSelect のチェックを切り替える |
| `/` | Select / MultiSelect の絞込み入力を始める |
| `Esc` | 絞込み入力を抜ける、または絞込みを解除する |
| `Ctrl+U` / `Ctrl+D` | リストを半ページずつ移動する |
| `Home` / `End` | リストの先頭／末尾へ移動する |
| `←` / `→` | Confirm の選択を切り替える |
| `y` / `n` | Confirm の true／false 側を選ぶ |
| `Alt+Enter` / `Ctrl+J` | Text で改行する |
| `Ctrl+E` | Input の候補補完、または有効化した Text の外部エディター |
| `Ctrl+C` | フォームを中断する |

絞込み中は `Enter` が検索の確定などへ使われる場面もあります。
`Space` を押しただけでは質問全体の回答完了にはならず、チェックの切替になります。
詳しくは [v2.0.3 の keymap.go](https://github.com/charmbracelet/huh/blob/v2.0.3/keymap.go) を参照してください。

<a id="troubleshooting"></a>
## 困った時：症状から確認する

| 症状 | 最初に確認すること |
| --- | --- |
| 関数が定義されていないと言われる | 別ファイルも含めて `go run .` で起動しているか |
| `Value` で型エラーになる | Input は `*string`、Confirm は `*bool`、MultiSelect は `*[]T` になっているか |
| メソッドチェーンの途中で構文エラーになる | 改行前の行末に `.` を置いているか |
| 初期値が表示されない・違う候補になる | 表示名ではなく候補の値を初期値にしているか |
| 入力例が保存されない | `Placeholder` は回答を設定しない。変数の初期値と区別しているか |
| MultiSelect で移動しただけでは選べない | Space でチェックし、Enter で進んでいるか |
| Text で Enter を押すと終わる | 改行に Alt+Enter または Ctrl+J を使っているか |
| 「いいえ」なのに生成された | Confirm の false を受けて、呼出し側で生成を止めているか |
| 固定 IP を無効にしても cfg に IP が残る | 質問を隠すだけでなく、回答と出力条件を再評価しているか |
| 国を変えても候補が更新されない | `OptionsFunc` の binding が変化を反映するポインターになっているか |
| 複数の入力欄が同じ回答になる | 同じ変数のポインターを全 Field へ渡していないか |
| map の回答を Value に渡せない | ローカル変数を経由して、正常終了後に書き戻しているか |
| テーマ指定で型エラーになる | v2.0.3 の `ThemeFunc` と `WithTheme` の型を確認したか |
| アクセシブルモードのメソッドがない | Field でなく Form の `WithAccessible` を使っているか |
| 実行中に画面が崩れる | Form.Run 中に標準出力へ別の文字を出していないか |
| 端末以外の実行で失敗する・待つ | 対話端末で実行しているか。アクセシブルモードも入力は必要 |

コード例を組み合わせる時は、「質問を表示する関数」「回答を検証する関数」「cfg を生成する関数」を分けて考えます。
まず一つの Field だけを動かし、それから Group、条件分岐、YAML 対応を追加すると原因を追いやすくなります。

## 参考資料

- [このプロジェクトの実装課題](implementation-exercises.md)
- [公式リポジトリ・v2.0.3](https://github.com/charmbracelet/huh/tree/v2.0.3)
- [API リファレンス・v2.0.3](https://pkg.go.dev/charm.land/huh/v2@v2.0.3)
- [v2.0.3 の Field インターフェース](https://github.com/charmbracelet/huh/blob/v2.0.3/huh.go)
- [v2 への移行ガイド](https://github.com/charmbracelet/huh/blob/v2.0.3/UPGRADE_GUIDE_V2.md)（コード例は手元の v2.0.3 の API と照合する）
- [Go の構造体](https://go.dev/tour/moretypes/2)
- [Go のポインター](https://go.dev/tour/moretypes/1)
- [Go のスライス](https://go.dev/tour/moretypes/7)
