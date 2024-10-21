---
sidebar_position: 2
---

# Helmet

The concept of the Helmet middleware is borrowed from the [Helmet](https://helmetjs.github.io/) library. It is used to secure your server by setting various HTTP headers.

Pine's implementation could still use some improvements and more on that will be updated in the next few weeks.

Nonetheless, the concept is simple and effective. The aim of this middleware is to set various HTTP headers to enhance the security of your server to prevent common attacks.

## Usage

```go
import (
  "github.com/BryanMwangi/pine"
  "github.com/BryanMwangi/pine/helmet"
  )

app := pine.New()

app.Use(helmet.New())
```

## Configuration

The Config struct for helmet has the following properties:

```go
type Config struct {
  // XSSProtection
  // Optional. Default value "0".
  XSSProtection string

  // ContentTypeNosniff
  // Optional. Default value "nosniff".
  ContentTypeNosniff string

  // XFrameOptions
  // Optional. Default value "SAMEORIGIN".
  // Possible values: "SAMEORIGIN", "DENY", "ALLOW-FROM uri"
  XFrameOptions string

  // HSTSMaxAge
  // Optional. Default value 0.
  HSTSMaxAge int

  // HSTSExcludeSubdomains
  // Optional. Default value false.
  HSTSIncludeSubdomains bool

  // ContentSecurityPolicy
  // Optional. Default value "".
  ContentSecurityPolicy string

  // CSPReportOnly
  // Optional. Default value false.
  CSPReportOnly bool

  // ReferrerPolicy
  // Optional. Default value "ReferrerPolicy".
  ReferrerPolicy string

  // Cross-Origin-Embedder-Policy
  // Optional. Default value "require-corp".
  CrossOriginEmbedderPolicy string

  // Cross-Origin-Opener-Policy
  // Optional. Default value "same-origin".
  CrossOriginOpenerPolicy string

  // Cross-Origin-Resource-Policy
  // Optional. Default value "same-origin".
  CrossOriginResourcePolicy string

  // Origin-Agent-Cluster
  // Optional. Default value "?1".
  OriginAgentCluster string

  // X-DNS-Prefetch-Control
  // Optional. Default value "off".
  XDNSPrefetchControl string

  // X-Download-Options
  // Optional. Default value "noopen".
  XDownloadOptions string

  // X-Permitted-Cross-Domain-Policies
  // Optional. Default value "none".
  XPermittedCrossDomain string
}
```

Here are the default values when setting up a new instance of helmet:

```go
  cfg := Config{
    XSSProtection:             "0",
    ContentTypeNosniff:        "nosniff",
    XFrameOptions:             "SAMEORIGIN",
    HSTSMaxAge:                0,
    HSTSIncludeSubdomains:     false,
    ContentSecurityPolicy:     "",
    ReferrerPolicy:            "ReferrerPolicy",
    CrossOriginEmbedderPolicy: "require-corp",
    CrossOriginOpenerPolicy:   "same-origin",
    CrossOriginResourcePolicy: "same-origin",
    OriginAgentCluster:        "?1",
    XDNSPrefetchControl:       "off",
    XDownloadOptions:          "noopen",
    XPermittedCrossDomain:     "none",
  }
```

## Properties

Helmet adds the following headers to the response:

| Header                            | Description                                                                                                                                                                                                              |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| X-XSS-Protection                  | Enables the browser's built-in protection against reflected XSS (cross-site scripting) attacks by preventing the browser from executing potentially malicious scripts.                                                   |
| X-Content-Type-Options            | Prevents MIME type sniffing attacks by instructing browsers to honor the server-specified Content-Type header, mitigating attacks that attempt to load files as a different type.                                        |
| X-Frame-Options                   | Helps prevent clickjacking attacks by controlling whether a page can be embedded in an iframe, typically blocking or restricting it to the same origin.                                                                  |
| Strict-Transport-Security         | Ensures that all future requests to your site are made over HTTPS, preventing man-in-the-middle attacks by enforcing secure connections.                                                                                 |
| Content-Security-Policy           | Protects against a range of attacks, such as cross-site scripting (XSS) and data injection, by specifying which content sources are considered trustworthy.                                                              |
| Cross-Origin-Embedder-Policy      | Prevents loading resources from other origins that do not explicitly grant permission, thereby protecting against cross-origin attacks like [Spectre](<https://en.wikipedia.org/wiki/Spectre_(security_vulnerability)>). |
| Cross-Origin-Opener-Policy        | Isolates browsing contexts by controlling whether a page can interact with pages from other origins, protecting against side-channel attacks like Spectre.                                                               |
| Cross-Origin-Resource-Policy      | Restricts which origins can load resources, preventing cross-origin data leaks and attacks such as [Spectre](<https://en.wikipedia.org/wiki/Spectre_(security_vulnerability)>).                                          |
| X-Permitted-Cross-Domain-Policies | Preventing data leakage through cross-domain requests.                                                                                                                                                                   |
| X-DNS-Prefetch-Control            | Controls whether browsers can prefetch DNS records, which can help mitigate privacy risks and performance issues associated with DNS prefetching.                                                                        |
| X-Download-Options                | Disable automatic execution of downloaded files, reducing exposure to malicious files.                                                                                                                                   |
| Origin-Agent-Cluster              | Helps mitigate side-channel attacks by isolating origins in separate agent clusters, preventing them from sharing certain resources like memory.                                                                         |
| Referrer-Policy                   | Controls the amount of referrer information sent along with requests, helping to prevent information leakage by specifying what data is passed to external sites.                                                        |

I will continue to do more research on ways I can improve Pine's security and I welcome any suggestions or corrections anyone may have to offer.
