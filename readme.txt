get-flickr-img

  flickrの画像をダウンロードするプログラム

*使い方

  Usage of ./get-flickr-img:
    -imgnum int     The number of images. (default 100)
    -output string  Output directory (default "output")

*オプション説明
  -imgnumでそれぞれのクエリに対して取得する画像数を決める。 デフォルトは100で500まで設定できる
  -outputは画像を保存するディレクトリで、予め作成しておく。 デフォルトはoutput

  APIキーは環境変数から取得するので.cshrcに'setenv FLICKR_API_KEY [APIキー]'をしていしておく必要がある。設定したらsource ~/.cshrcコマンドで設定を反映させる。
  もしくはenv FLICKR_API_KEY=[APIキー] get-flickr- img

*コマンド例

  ./get-flickr-img < queries.txt

  queries.txtは以下のようにして作成する。クエリ文字列を箇条書きする。

  sea
  mountain

