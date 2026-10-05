## パッケージ

利用する予定のパッケージは以下の通りです。

| YAML読み込み           | `go.yaml.in/yaml/v4`                               |
| ---------------------- | -------------------------------------------------- |
| 質問・入出力           | `bufio`, `fmt`, `os`, `flag`, `strconv`, `strings` |
| パスワード入力の非表示 | `golang.org/x/term`                                |
| preseedテキスト生成    | `text/template`                                    |
| YAMLをexeに埋め込み    | `embed`                                            |
| パスワードハッシュ     | `github.com/sergeymakinen/go-crypt/sha512`         |