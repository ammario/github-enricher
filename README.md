# github-enricher

`github-enricher` enriches GitHub data.

It accepts CSV-formatted input and emits CSV-formatted output. It's a good sidekick to [GitHub's official
bigquery dataset](https://cloud.google.com/blog/topics/public-datasets/github-on-bigquery-analyze-all-the-open-source-code), which redacts email addresses.

## Cache

`github-enricher` requires Redis as caching layer. This is absolutely essential for performance, as cacheless enrichment is bottlenecked by GitHub API limits and expensive clones.

The `REDIS_ADDR` and `REDIS_PASSWORD` environment variables are used to configure the cache client.

The cache allows for fast incremental enrichment.

## Columns

| Name      | Description                              | Dependencies       |
| --------- | ---------------------------------------- | ------------------ |
| repo_name | Repository name, e.g `torvalds/linux`    | Cannot be enriched |
| ref       | e.g   `master`                           | Cannot be enriched |
| email     | user's email as captured from commit     | repo_name, ref     |
| name      | user's full name as captured from commit | repo_name, ref     |

Any unrecogized columns are passed through verbatim.

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
