# Changelog

## [v0.18.1](https://github.com/x-motemen/blogsync/compare/v0.18.0...v0.18.1) - 2023-10-28
- suppress git error logs by @Songmu in https://github.com/x-motemen/blogsync/pull/119

## [v0.18.0](https://github.com/x-motemen/blogsync/compare/v0.17.0...v0.18.0) - 2023-10-28
- Drop testify dependency by @Songmu in https://github.com/x-motemen/blogsync/pull/113
- Output debug logs for API access by @Songmu in https://github.com/x-motemen/blogsync/pull/115
- use git log to determine when a file was last updated by @Songmu in https://github.com/x-motemen/blogsync/pull/118
- add fetch subcommand by @Songmu in https://github.com/x-motemen/blogsync/pull/116

## [v0.17.0](https://github.com/x-motemen/blogsync/compare/v0.16.1...v0.17.0) - 2023-10-22
- Add PreviewURL field by @Mahito in https://github.com/x-motemen/blogsync/pull/104
- [incompatible] change log output dest to stderr by @Songmu in https://github.com/x-motemen/blogsync/pull/106
- output updated files to stdout by @Songmu in https://github.com/x-motemen/blogsync/pull/107
- add "--no-drafts" and "--only-drafts" flags to pull subcommand for drafts management by @Songmu in https://github.com/x-motemen/blogsync/pull/110

## [v0.16.1](https://github.com/x-motemen/blogsync/compare/v0.16.0...v0.16.1) - 2023-10-14
- 記事更新時に更新元ファイルのentryHeaderのURLが未設定の場合も更新ができるように修正  by @halkt in https://github.com/x-motemen/blogsync/pull/99

## [v0.16.0](https://github.com/x-motemen/blogsync/compare/v0.15.0...v0.16.0) - 2023-10-14
- add doc directory by @Songmu in https://github.com/x-motemen/blogsync/pull/92
- post new entry when the pushed entry has no EditURL by @Songmu in https://github.com/x-motemen/blogsync/pull/91
- set the updated to nil when the entry is still draft by @Songmu in https://github.com/x-motemen/blogsync/pull/95

## [v0.15.0](https://github.com/x-motemen/blogsync/compare/v0.14.2...v0.15.0) - 2023-10-13
- [breaking change] normalize localRoot to absolute path by @Songmu in https://github.com/x-motemen/blogsync/pull/87
- refactoring around config loading by @Songmu in https://github.com/x-motemen/blogsync/pull/89
- expand tilde in LocalPath with home directory by @Songmu in https://github.com/x-motemen/blogsync/pull/90

## [v0.14.2](https://github.com/x-motemen/blogsync/compare/v0.14.1...v0.14.2) - 2023-10-09
- [bugfix] ignore errors in the case of fetching fixed pages by @Songmu in https://github.com/x-motemen/blogsync/pull/85

## [v0.14.1](https://github.com/x-motemen/blogsync/compare/v0.14.0...v0.14.1) - 2023-10-09
- update GitHub Actions workflows by @Songmu in https://github.com/x-motemen/blogsync/pull/81
- remove old file that is no longer used due to changing the path of the entry after updating it by @Songmu in https://github.com/x-motemen/blogsync/pull/83

## [v0.14.0](https://github.com/x-motemen/blogsync/compare/v0.13.6...v0.14.0) - 2023-10-08
- support fixed pages by @Songmu in https://github.com/x-motemen/blogsync/pull/78
- Go 1.21 and update deps by @Songmu in https://github.com/x-motemen/blogsync/pull/80

## [v0.13.6](https://github.com/x-motemen/blogsync/compare/v0.13.5...v0.13.6) - 2023-10-08
- 実際のブログを使ったE2Eテスト by @motemen in https://github.com/x-motemen/blogsync/pull/75
- s/remoteRoot/blogID/ by @Songmu in https://github.com/x-motemen/blogsync/pull/77

## [v0.13.5](https://github.com/x-motemen/blogsync/compare/v0.13.4...v0.13.5) - 2023-04-16
- run tests on windows on CI by @Songmu in https://github.com/x-motemen/blogsync/pull/72
- update deps by @Songmu in https://github.com/x-motemen/blogsync/pull/74

## [v0.13.4](https://github.com/x-motemen/blogsync/compare/v0.13.3...v0.13.4) - 2023-04-16
- Enable stdin on windows by @ongaeshi in https://github.com/x-motemen/blogsync/pull/70

## [v0.13.3](https://github.com/x-motemen/blogsync/compare/v0.13.2...v0.13.3) - 2023-04-14
- adjust release/action.yml by @Songmu in https://github.com/x-motemen/blogsync/pull/68

## [v0.13.2](https://github.com/x-motemen/blogsync/compare/v0.13.1...v0.13.2) - 2023-04-14
- adjust GitHub Actions for action-update-semver by @Songmu in https://github.com/x-motemen/blogsync/pull/66

## [v0.13.1](https://github.com/x-motemen/blogsync/compare/v0.13.0...v0.13.1) - 2023-04-14
- add action.yml to support custom GitHub Actions by @Songmu in https://github.com/x-motemen/blogsync/pull/64
- Marketplaceに公開するためのアクションの作成 by @theoremoon in https://github.com/x-motemen/blogsync/pull/63

## [v0.13.0](https://github.com/x-motemen/blogsync/compare/v0.12.3...v0.13.0) - 2023-01-25
- AtomPub API のエンドポイントにownerを設定できるように修正 by @halkt in https://github.com/x-motemen/blogsync/pull/58

## [v0.12.3](https://github.com/x-motemen/blogsync/compare/v0.12.2...v0.12.3) - 2022-12-22
- chore: build assets for Windows by @suzuki-shunsuke in https://github.com/x-motemen/blogsync/pull/56

## [v0.12.2](https://github.com/x-motemen/blogsync/compare/v0.12.1...v0.12.2) - 2022-12-21
- fix nil dereference when without default config by @Azuki-bar in https://github.com/x-motemen/blogsync/pull/51
- update release related stuffs by @Songmu in https://github.com/x-motemen/blogsync/pull/52
- update deps by @Songmu in https://github.com/x-motemen/blogsync/pull/53

## [v0.12.1](https://github.com/x-motemen/blogsync/compare/v0.12.0...v0.12.1) (2022-08-12)

* releng related file changes [#50](https://github.com/x-motemen/blogsync/pull/50) ([Songmu](https://github.com/Songmu))
* Update to Go 1.19 [#49](https://github.com/x-motemen/blogsync/pull/49) ([w-haibara](https://github.com/w-haibara))

## [v0.12.0](https://github.com/x-motemen/blogsync/compare/v0.11.0...v0.12.0) (2020-07-09)

* username and password can read from environment variable  [#48](https://github.com/x-motemen/blogsync/pull/48) ([nabeo](https://github.com/nabeo))
* change urls contained in documents to the correct repository [#45](https://github.com/x-motemen/blogsync/pull/45) ([lufia](https://github.com/lufia))

## [v0.11.0](https://github.com/x-motemen/blogsync/compare/v0.10.3...v0.11.0) (2020-01-24)

* change module name to github.com/x-motemen/blogsync [#44](https://github.com/x-motemen/blogsync/pull/44) ([Songmu](https://github.com/Songmu))

## [v0.10.3](https://github.com/x-motemen/blogsync/compare/v0.10.2...v0.10.3) (2020-01-05)

* introduce GitHub actions [#43](https://github.com/x-motemen/blogsync/pull/43) ([Songmu](https://github.com/Songmu))
* push subcommand pushes entr"ies" [#42](https://github.com/x-motemen/blogsync/pull/42) ([kiririmode](https://github.com/kiririmode))
* update urfave/cli to v2 [#41](https://github.com/x-motemen/blogsync/pull/41) ([Songmu](https://github.com/Songmu))
* follow urfave/cli's change [#40](https://github.com/x-motemen/blogsync/pull/40) ([Songmu](https://github.com/Songmu))

## [v0.10.2](https://github.com/x-motemen/blogsync/compare/v0.10.1...v0.10.2) (2019-12-12)

* update deps [#38](https://github.com/x-motemen/blogsync/pull/38) ([Songmu](https://github.com/Songmu))
* [maint] update toolchains [#37](https://github.com/x-motemen/blogsync/pull/37) ([Songmu](https://github.com/Songmu))
* remove github.com/mitchellh/go-homedir dependency [#36](https://github.com/x-motemen/blogsync/pull/36) ([itchyny](https://github.com/itchyny))
* Fix: Typo [#35](https://github.com/x-motemen/blogsync/pull/35) ([codeout](https://github.com/codeout))

## [v0.10.1](https://github.com/x-motemen/blogsync/compare/v0.10.0...v0.10.1) (2019-02-23)

* introduce go modules and Songmu/godzil [#34](https://github.com/x-motemen/blogsync/pull/34) ([Songmu](https://github.com/Songmu))

## [v0.10.0](https://github.com/x-motemen/blogsync/compare/v0.9.1...v0.10.0) (2018-10-26)

* update deps [#32](https://github.com/x-motemen/blogsync/pull/32) ([Songmu](https://github.com/Songmu))
* Add list command [#29](https://github.com/x-motemen/blogsync/pull/29) ([autopp](https://github.com/autopp))
* fix path of golint [#31](https://github.com/x-motemen/blogsync/pull/31) ([autopp](https://github.com/autopp))
* enhance testing around loading configuration [#27](https://github.com/x-motemen/blogsync/pull/27) ([Songmu](https://github.com/Songmu))

## [v0.9.1](https://github.com/x-motemen/blogsync/compare/v0.9.0...v0.9.1) (2018-07-29)

* allow nil config, which config values are cascaded from the default config [#26](https://github.com/x-motemen/blogsync/pull/26) ([Songmu](https://github.com/Songmu))

## [v0.9.0](https://github.com/x-motemen/blogsync/compare/v0.1.0...v0.9.0) (2018-07-29)

* update deps [#25](https://github.com/x-motemen/blogsync/pull/25) ([Songmu](https://github.com/Songmu))
* fill the default values if the configuration value is empty when getting blogConfig [#24](https://github.com/x-motemen/blogsync/pull/24) ([Songmu](https://github.com/Songmu))
* mergeBlogConfig [#23](https://github.com/x-motemen/blogsync/pull/23) ([Songmu](https://github.com/Songmu))
* add config.OmitDomain [#22](https://github.com/x-motemen/blogsync/pull/22) ([Songmu](https://github.com/Songmu))
* refine `commandPush` [#21](https://github.com/x-motemen/blogsync/pull/21) ([Songmu](https://github.com/Songmu))

## [v0.1.0](https://github.com/x-motemen/blogsync/compare/v0.0.1...v0.1.0) (2018-05-08)

* introduce dep [#19](https://github.com/x-motemen/blogsync/pull/19) ([Songmu](https://github.com/Songmu))
* drop entryTime and use `*time.Time` directly [#18](https://github.com/x-motemen/blogsync/pull/18) ([Songmu](https://github.com/Songmu))

## [v0.0.1](https://github.com/x-motemen/blogsync/compare/acc350d...v0.0.1) (2017-08-16)

* add releng related files [#16](https://github.com/x-motemen/blogsync/pull/16) ([Songmu](https://github.com/Songmu))
* add Makefilea and .travis.yml [#15](https://github.com/x-motemen/blogsync/pull/15) ([Songmu](https://github.com/Songmu))
* Eradicate linter's point [#14](https://github.com/x-motemen/blogsync/pull/14) ([Songmu](https://github.com/Songmu))
* Maint [#13](https://github.com/x-motemen/blogsync/pull/13) ([Songmu](https://github.com/Songmu))
* Merge config files [#12](https://github.com/x-motemen/blogsync/pull/12) ([Songmu](https://github.com/Songmu))
* eradicate dieIf [#11](https://github.com/x-motemen/blogsync/pull/11) ([Songmu](https://github.com/Songmu))
* maint deps (codegangsta/cli moved to urfave/cli) [#10](https://github.com/x-motemen/blogsync/pull/10) ([Songmu](https://github.com/Songmu))
* change config file searching way [#9](https://github.com/x-motemen/blogsync/pull/9) ([Songmu](https://github.com/Songmu))
* if there is not config.yml in HOME, look up current dir [#8](https://github.com/x-motemen/blogsync/pull/8) ([t-mrt](https://github.com/t-mrt))
* Support entry categories [#7](https://github.com/x-motemen/blogsync/pull/7) ([tarao](https://github.com/tarao))
* supoprt custom url with --custom-path option [#5](https://github.com/x-motemen/blogsync/pull/5) ([Songmu](https://github.com/Songmu))
* fix for the case Entry.Date is nil [#4](https://github.com/x-motemen/blogsync/pull/4) ([motemen](https://github.com/motemen))
* Frontmatter [#2](https://github.com/x-motemen/blogsync/pull/2) ([Songmu](https://github.com/Songmu))
* Using `RemoteRoot` in `LocalPath` instead of `URL.Host` [#1](https://github.com/x-motemen/blogsync/pull/1) ([Songmu](https://github.com/Songmu))
