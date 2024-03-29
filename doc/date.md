# はてなブログAtomPubの日時フィールドに関する振る舞いの調査

AtomPubのエントリーにはいくつかの日時フィールドがあるが、それぞれ何を表しているか実際に振る舞いを確かめながら調べた。

## フィールドの対応

AtomPub上のデータと、公開リソース内の対応は以下のようになっている。Atom Pub上のupdatedが公開リソース上のpublishに対応していることや、AtomPub上のapp:editedが公開リソース上のupdated等に対応しているのが混乱の元。

| AtomPub    | Atom Feed | HTML          | RSS |
| ---------  | --------- | ------------- |---- |
| updated    | published | datePublished | pubDate |
| published  | -         | -             | -       |
| app:edited | updated   | dateModified  | -       |

投稿者が指定可能な「投稿日時」がupdatedであり、これを「公開日時」的に扱う仕様になっている。実際の本当の公開日時はAtomPubのpublishedが対応するが、これは公開リソース上では特に使われていない。また投稿者が指定可能な日時はこの「投稿日時(=updated)」のみである。

### わかったこと
- updated
    - 記事編集画面の「投稿日時」に対応
    - ブログ記事上に表示されている日時もこれになる
    - 投稿者が指定可能
        - 投稿者が任意に設定可能な「公開日」的な位置づけになっている
        - 投稿者が指定可能な時刻は逆にこれだけ
    - 指定しない場合は以下のような挙動になる
        - 指定しない場合とはブログ編集画面で投稿日時を空白にした場合や、AtomPubでupdatedをしていない場合
        - 記事公開時にその時点の時刻が使われる
            - ブログ編集画面サイドバーの「投稿日時」の値も設定される
        - 下書き状態で指定しない場合 には、公開時と異なり、恐らくはてなブログ内部のDBではNULLになる
            - ブログ編集画面サイドバーの投稿日時が空欄のままであるため
                - ![](http://songmu.github.io/images/ghzo/23-1014-2157-441f8c813f199a62.png)
            - その場合であってもAtomPubのレスポンスのupdatedはNULLにはなっておらず、app:editedと同じ時刻が埋まっている
            - 下書き状態であっても一度明示的に投稿日時を指定すると空欄に戻す方法は無い
                - ブログ編集画面で空欄に戻したりAtomPubでnullに設定しようとしても無理だった
                - 編集画面も空欄ではなくなり、app:editedと同じ値に自動設定されることもなくなる
    - 予約投稿でもupdatedが指定した日時に設定され、Draft扱いのまま
        - 未来の日時に投稿日時を指定した下書きとの区別はデータ上は不可能
        - 予約投稿が実際に行われた時点で、publishedとapp:editedも同じ時刻に更新される
- published
    - 実際の公開日時であり投稿者でも変更や指定はできない
        - 記事公開後は下書きに戻されない限りは不変
    - このデータはAtomPub以外で露出することはない
        - つまりブログ閲覧者が見る方法はない
    - 下書き状態では下書きが作られた日時に暫定的に設定される
        - 公開時に公開日時に変更される
    - 下書きに戻して再公開した場合には、再公開日時に更新される
        - 下書きに戻した時点では変更されない
- app:edited
    - 実際の最終更新日時であり投稿者でも変更や指定はできない
    - Atom Feedの updated などにこの値が入っている
    - おそらく、はてなブログ内部で非同期に記事が更新された場合にも更新される
        - fotolife URLの非同期書き換えなどの場合

## その他
- 下書き時にCustumPathを指定していない場合、公開予定URLの場所が変更になるがこれはどういうルール？
    - → app:updated 基準になっている
