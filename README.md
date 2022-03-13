# github-enricher

This tool helps enriches a list of GitHub usernames into email addresses and other profile data.

`github-enricher` is CSV oriented. It accepts CSV-formatted input and emits CSV-formatted output.

## Redis Cache

`github-enricher` relies on Redis as a caching layer.

## Columns

| Name      | Description                              | Dependencies       |
| --------- | ---------------------------------------- | ------------------ |
| repo_name | Repository name, e.g `torvalds/linux`    | Cannot be enriched |
| ref       | e.g   `master`                           | Cannot be enriched |
| email     | user's email as captured from commit     | repo_name, ref     |
| name      | user's full name as captured from commit | repo_name, ref     |

Any unrecogized columns are passed through verbatim.

The first line of input and output is always a header.

## Basic Usage

```bash
$ echo "name,repo_name,ref
Chromium LUCI CQ,chromium/chromium,c33d4dbfd275d5659cc2c79cbec75810ae4bdd37
TypeScript Bot,kitsonk/TypeScript,2d80473c781818b1712c6106fd8b1faea59d25ae
GitHub,Azure/azure-sdk-for-python,23decbe4b61626b6a37f1f23dcf18514a2f445a5
gVisor bot,google/gvisor,1fcaa119a53ada26ade9fb1405cd593204699adc
" | go run github.com/ammario/github-enricher
email,orgname
ammar@ammar.io,@coderhq
```
