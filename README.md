# github-enricher

`github-enricher` enriches GitHub data.

It accepts CSV-formatted input and emits CSV-formatted output. It's a good sidekick to [GitHub's official
bigquery dataset](https://cloud.google.com/blog/topics/public-datasets/github-on-bigquery-analyze-all-the-open-source-code), which redacts email addresses.

## Caching

`github-enricher` is designed for fast incremental enrichment. Thus, it requires on Redis and the filesystem
for caching.

### Redis

The `REDIS_ADDR` and `REDIS_PASSWORD` environment variables are used to configure the cache client.

Make sure your redis is persisting (e.g `save 60 1` in your redis.conf).

### Filesystem

Even though repos are shallow cloned, it can take minutes to retrieve a commit from a large repo. All repos
are cloned to the `github-enricher` folder in your OS tempdir.

## Columns

| Name      | Description                                                                       | Dependencies       |
| --------- | --------------------------------------------------------------------------------- | ------------------ |
| repo_name | Repository name, e.g `torvalds/linux`                                             | Cannot be enriched |
| ref       | e.g `master`                                                                      | Cannot be enriched |
| email     | user's email as captured from commit                                              | repo_name, ref     |
| name      | user's full name as captured from commit                                          | repo_name, ref     |
| username  | GitHub login name                                                                 | repo_name, ref     |
| gender    | probable gender from first name [hstove/gender](https://github.com/hstove/gender) | name               |
| firstname | first word in name                                                                | name               |
| lastname  | last word in name                                                                 | name               |

All unrecogized columns are passed through verbatim.

The first line of the input and output is always a header.

## Example

This examples enriches email addresses from a list of commits. `name` is passed through untouched.

input.csv:

```csv
TensorFlower Gardener,keras-team/keras,9b14e16b8cc93abcc21355115a7a18c34d385281
Chromium LUCI CQ,chromium/chromium,c33d4dbfd275d5659cc2c79cbec75810ae4bdd37
TypeScript Bot,kitsonk/TypeScript,2d80473c781818b1712c6106fd8b1faea59d25ae
GitHub,Azure/azure-sdk-for-python,23decbe4b61626b6a37f1f23dcf18514a2f445a5
```

shell invokation:

```bash
$ go run github.com/ammario/github-enricher < input.csv
name,repo_name,commit,email
TensorFlower Gardener,keras-team/keras,9b14e16b8cc93abcc21355115a7a18c34d385281,mattdangerw@google.com
Chromium LUCI CQ,chromium/chromium,c33d4dbfd275d5659cc2c79cbec75810ae4bdd37,ppz@chromium.org
TypeScript Bot,kitsonk/TypeScript,2d80473c781818b1712c6106fd8b1faea59d25ae,typescriptbot@microsoft.com
```
