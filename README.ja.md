# go-psnr

Pure Go による高速な PSNR（ピーク信号対雑音比）計算ライブラリ

[![Go Reference](https://pkg.go.dev/badge/github.com/ideamans/go-psnr.svg)](https://pkg.go.dev/github.com/ideamans/go-psnr)
[![CI](https://github.com/ideamans/go-psnr/actions/workflows/ci.yml/badge.svg)](https://github.com/ideamans/go-psnr/actions/workflows/ci.yml)

## 特徴

- **高速**: 整数演算と最適化されたアルゴリズムを使用
- **互換性**: ImageMagick と 2%以内の誤差で一致
- **Pure Go**: CGo に依存せず、Go が動作する環境ならどこでも実行可能
- **シンプルな API**: ファイルパスまたはバイトスライスで簡単に使用可能
- **フォーマットサポート**: JPEG および PNG 形式に対応

## インストール

```bash
go get github.com/ideamans/go-psnr
```

## 使用方法

### 基本的な使い方

```go
package main

import (
    "fmt"
    "log"
    "github.com/ideamans/go-psnr/psnr"
)

func main() {
    // ファイルパスからPSNRを計算
    value, err := psnr.ComputeFiles("image1.jpg", "image2.jpg")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("PSNR: %.2f dB\n", value)
}
```

### バイトスライスを使用する場合

```go
// 画像をバイトスライスに読み込む
data1, _ := os.ReadFile("image1.png")
data2, _ := os.ReadFile("image2.png")

// PSNRを計算
value, err := psnr.Compute(data1, data2)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("PSNR: %.2f dB\n", value)
```

## パフォーマンス

このパッケージは以下の最適化を使用しています：

- MSE 計算における整数演算
- 一般的な画像形式（RGBA、NRGBA、YCbCr）用の高速パス
- 最適化されたアルファチャンネル検出
- サポートされた形式での直接ピクセルバッファアクセス

## ImageMagick との互換性

このパッケージは、ImageMagick（libjpeg 使用）と互換性のある PSNR 値を生成するよう設計されており、通常 ImageMagick の計算結果と 2%以内の誤差で一致します。わずかな差異の原因：

- YCbCr→RGB 変換時の丸め処理の違い
- IDCT（逆離散コサイン変換）の実装差
- JPEG デコーダーの実装差（Go の Image/jpeg vs libjpeg）

この精度により、多くのアプリケーションで ImageMagick の PSNR 計算の代替として使用できます。

## 動作要件

- Go 1.22 以降

## ライセンス

MIT License - 詳細は LICENSE ファイルを参照してください。
