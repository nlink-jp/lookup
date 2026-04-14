# lookup: JSONデータエンリッチメントCLIツール

`lookup` は、Splunkの `lookup` コマンドにインスパイアされたコマンドラインユーティリティです。外部データソース（CSVやJSONファイル）の値に基づいてフィールドを追加することで、JSONデータストリームをエンリッチします。

標準入力からJSONオブジェクト（JSON配列またはJSON Lines形式）を読み込み、設定されたルールに基づいてルックアップを実行し、エンリッチされたJSONオブジェクトを標準出力に出力します。

---

## 特徴

- **複数のデータソース**: **CSV** または **JSON** ファイルをルックアップテーブルとして使用
- **高度なマッチングメソッド**:
  - `exact`: 大文字小文字を区別する/しない完全一致
  - `wildcard`: グロブスタイルのワイルドカードマッチング（例: `bot-*`）
  - `regex`: 正規表現による強力なマッチング
  - `cidr`: IPアドレスとCIDRブロックの照合（例: `10.0.0.0/8`）
- **柔軟な設定**: JSON設定ファイルでルックアップロジックとデータを分離
- **組み込みDNSルックアップ**: 正引き（`A`レコード）または逆引き（`PTR`レコード）をネイティブ機能として実行
  - カスタムDNSサーバーの指定も可能
- **柔軟なフィールドマッピング**: 直感的な構文でマッチフィールドと出力フィールド名を制御
- **複数の入力形式**: **JSON配列** と **JSON Lines (JSONL)** の両方を自動検出して処理
- **クロスプラットフォーム**: Go製の単一バイナリで、Linux、macOS、Windowsで動作

---

## インストール

macOS、Windows、Linux向けのコンパイル済みバイナリは[リリースページ](https://github.com/nlink-jp/lookup/releases)から入手できます。

---

## 設定

### 設定ファイル (`config.json`)

```json
{
  "data_source": "./path/to/your/data.csv",
  "matchers": [
    {
      "input_field": "field_from_stdin",
      "lookup_field": "column_in_data_source",
      "method": "exact",
      "case_sensitive": false
    }
  ]
}
```

- **`data_source`**: ルックアップデータファイルへの相対パスまたは絶対パス
- **`matchers`**: マッチングルールの配列
  - **`input_field`**: このルックアップルールの名前。`-m` フラグで参照
  - **`lookup_field`**: データソースファイルのカラム/キー名
  - **`method`**: `"exact"`（デフォルト）、`"wildcard"`、`"regex"`、`"cidr"`
  - **`case_sensitive`**: `true`で大文字小文字を区別。デフォルトは`false`

### 設定ヘルパー (`generate-config`)

```sh
./lookup generate-config -file <データファイルへのパス>
```

データファイルのヘッダー/キーをスキャンして設定テンプレートを生成します。

---

## 使い方

```sh
cat input.json | ./lookup -c <config.json> -m "<マッピングルール>"
```

### コマンドラインフラグ

| フラグ | 説明 | 必須 |
|--------|------|------|
| `-c <パス>` | JSON設定ファイルのパス | はい |
| `-m <文字列>` | マッピングルール文字列 | はい |
| `--dns` | DNSルックアップモードを有効化。`-c`は無視される | いいえ |
| `--dns-server` | カスタムDNSサーバーアドレス（例: `8.8.8.8`） | いいえ |

### マッピング構文

```
"CONFIG_REF_FIELD as INPUT_FIELD [OUTPUT original_name1 as new_name1, original_name2]"
```

- **`CONFIG_REF_FIELD as INPUT_FIELD`**: 必須。マッチャーの`input_field`と入力JSONのフィールド名を紐付け
- **`OUTPUT ...`**: 任意。出力フィールドの選択とリネーム。省略時は全フィールドが追加される

### 使用例

#### 1. 完全一致（大文字小文字無視）

```sh
cat input.jsonl | ./lookup -c config.json -m "user as user OUTPUT department as dept, role"
```

#### 2. CIDRマッチ

```sh
cat input.jsonl | ./lookup -c config.json -m "ip_lookup as client_ip"
```

#### 3. DNSルックアップ

```sh
echo '{"ip":"8.8.8.8"}' | ./lookup --dns -m "dns as ip OUTPUT hostname"
```

---

## ビルド

```sh
make build        # 現在のプラットフォーム向けにビルド
make build-all    # 全プラットフォーム向けにクロスコンパイル
make test         # テスト実行
make package      # ビルド + zipアーカイブ作成
```

ビルド成果物は `dist/` ディレクトリに配置されます。

---

## ドキュメント

- [Architecture](docs/en/architecture.md) — モジュール構成、データフロー、テスト戦略
- [アーキテクチャ](docs/ja/architecture.ja.md) — 日本語版

---

## ライセンス

このプロジェクトは [MITライセンス](LICENSE) の下で公開されています。
