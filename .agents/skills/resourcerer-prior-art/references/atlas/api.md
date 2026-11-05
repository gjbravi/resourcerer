---
updatedAt: 2026-07-08T10:43:33.000Z
---

# API

### What is Atlas's API rate limit?

The rate limit depends on the endpoint you use.

The information about rate limits is available as an annotation for the path: `nginx.ingress.kubernetes.io/limit-rps`

Atlas uses the [Cloudflare](https://developers.cloudflare.com/fundamentals/api/reference/limits/#:~:text=The%20global%20rate%20limit%20for,receiving%20a%20HTTP%20429%20response.) default rate limiter before traffic arrives on our platform.

***

### How can I use Atlas's API to fetch prices for all namespaces between Date A and Date B?

One way to accomplish this is by creating an [allocation group](https://docs.atlas.example/docs/allocation-groups) based on the specific namespaces or all namespaces you have for the given cluster.

Next, you need to generate a cost allocation group summary for the specific data range.