# go-fast-psnr

Pure Goによる高速なPSNR（ピーク信号対雑音比）計算ライブラリ

[![Go Reference](https://pkg.go.dev/badge/github.com/ideamans/go-fast-psnr.svg)](https://pkg.go.dev/github.com/ideamans/go-fast-psnr)
[![CI](https://github.com/ideamans/go-fast-psnr/actions/workflows/ci.yml/badge.svg)](https://github.com/ideamans/go-fast-psnr/actions/workflows/ci.yml)

## 特徴

- **高速**: 整数演算と最適化されたアルゴリズムを使用
- **互換性**: ImageMagickと2%以内の誤差で一致
- **Pure Go**: CGoに依存せず、Goが動作する環境ならどこでも実行可能
- **シンプルなAPI**: ファイルパスまたはバイトスライスで簡単に使用可能
- **フォーマットサポート**: JPEGおよびPNG形式に対応

## インストール

```bash
go get github.com/ideamans/go-fast-psnr
```

## 使用方法

### 基本的な使い方

```go
package main

import (
    "fmt"
    "log"
    "github.com/ideamans/go-fast-psnr/psnr"
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

- MSE計算における整数演算
- 一般的な画像形式（RGBA、NRGBA、YCbCr）用の高速パス
- 最適化されたアルファチャンネル検出
- サポートされた形式での直接ピクセルバッファアクセス

## ImageMagickとの互換性

このパッケージは、ImageMagick（libjpeg使用）と互換性のあるPSNR値を生成するよう設計されており、通常ImageMagickの計算結果と2%以内の誤差で一致します。わずかな差異の原因：

- YCbCr→RGB変換時の丸め処理の違い
- IDCT（逆離散コサイン変換）の実装差
- JPEGデコーダーの実装差（GoのImage/jpeg vs libjpeg）

この精度により、多くのアプリケーションでImageMagickのPSNR計算の代替として使用できます。

## 動作要件

- Go 1.22以降

## ライセンス

MIT License - 詳細はLICENSEファイルを参照してください。