# Robots.txt Extender

## Motivation

I needed to exclude the `/cdn-cgi/` subdirectory from being crawled by search engines. This directory is used by
Cloudflare and produces warnings in Google Search Console. [The official Cloudflare documentation recommends
updating the `robots.txt` file to include `Disallow: /cdn-cgi/`.](https://developers.cloudflare.com/fundamentals/reference/cdn-cgi-endpoint/)

Unfortunately, [Ghost](https://ghost.org/) (the blogging platform I use) does not allow directly editing
the `robots.txt` file and delegates this to the theme. This would require me to re-compile the theme
every time I wanted to update the `robots.txt` file. I just wanted a more flexible solution (and wanted
to play with Go, GitHub Actions, Go's Prometheus client, Go's structured logging, etc. -- this probably
being the more driving factor 😅).

In my setup, I can just configure my reverse proxy to forward requests to `/robots.txt` to this service instead
of the original Ghost deployment.

## Configuration

The following configuration is supported via environment variables:

| **Environment Variable**         | **Description**                                                                                                                                                                                                                                                              | **Default value**       | **Required** |
|----------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------|--------------|
| `PORT`                           | The port the application should listen to.                                                                                                                                                                                                                                   | `80`                    | No           |
| `ORIGINAL_ROBOTS_URL`            | The URL to fetch the original `robots.txt` from.                                                                                                                                                                                                                             | -                       | Yes          |
| `TIMEOUT_ROBOTS_REQUEST_SECONDS` | The timeout to use when requesting original `robots.txt` file from `ORIGINAL_ROBOTS_URL` in Go's duration format.                                                                                                                                                            | 5s                      | No           |
| `ADDITIONAL_ROBOTS_FILE`         | The file whose content to add to the original `robots.txt`. If you want introduce a new section (for example, for a separate `User-agent`, just add an empty line of the beginning of your file). The path can be relative to the current working directory of this service. | `additional_robots.txt` | No           |
| `ENDPOINT`                       | The endpoint to host the `robots.txt` file under from the root of the service. In case you need / want to find other ways to host the file.                                                                                                                                  | `robots.txt`            | No           |
| `INCLUDE_ORIGINAL_HEADERS`       | This allows keeping the original headers the underlying service is using for the `robots.txt` file. The following values are accepted: `1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False`.                                                                            | `true`                  | No           |
| `LOG_LEVEL`                      | The default log level to use.                                                                                                                                                                                                                                                | `info`                  | No           |

The Docker image includes a basic `additional_robots.txt` file with `Disallow: /cdn-cgi/`.

## Errors

In case of errors with the underlying endpoint, this service will return a `502 Bad Gateway`.

## Metrics

This service provides a `/metrics` endpoint for Prometheus with the following metrics:

| **Metric**                      | **Description**                                                                                                                                                                                                                                                                                    |
|---------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| total_robots_txt_requests       | The total number of requests for the robots.txt file provided by this service.                                                                                                                                                                                                                     |
| total_robots_txt_request_errors | The total number of errors when serving the robots.txt file. The reason for this is mostly the underlying robots.txt not being served properly or network issues. Not all errors may have been reflected back to the client, for example if an error occurred after the response has been started. |
