# twpicd

一键下载指定用户的上传的所有图片（不包括转推的）

# config

到 https://developer.twitter.com/ 申请 api key

设置以下环境变量为对应的值

- `TWITTER_CONSUMER_KEY`
- `TWITTER_CONSUMER_SECRET`
- `TWITTER_ACCESS_TOKEN`
- `TWITTER_ACCESS_SECRET`

# usage


`twpicd -name <username>`

# proxy

设置环境变量 `HTTPS_PROXY`


# other

官方 api 限制最多只能拿到用户最近的3200条推文
因此可能下载的图片不全