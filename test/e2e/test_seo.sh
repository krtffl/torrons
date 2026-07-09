#!/usr/bin/env bash
# SEO surface — CONFIRMED live bug. chi's middleware.URLFormat rewrites the
# routing path (rctx.RoutePath) to strip a trailing ".ext" before the router
# matches. The three SEO routes are registered WITH their extension
# (/robots.txt, /sitemap.xml, /llms.txt), so:
#   - request "/sitemap.xml" -> RoutePath stripped to "/sitemap" -> no route "/sitemap" -> 404
#   - request "/sitemap"     -> RoutePath stays "/sitemap"       -> no route "/sitemap" -> 404
# i.e. the robotsTxt / sitemapXML / llmsTxt handlers are DEAD CODE: unreachable
# at every URL. The ".png" image routes dodge this by registering WITHOUT the
# extension, so both "/share/card" and "/share/card.png" resolve to one route.
section "SEO — sitemap / robots / llms unreachable (URLFormat extension-stripping bug)"

# Public URLs a crawler hits: all 404 today. Want 200 once fixed (register the
# routes without the extension, or exempt these paths from URLFormat).
xfail_status SEO01 "robots.txt 404 (want 200)"   404 200 GET /robots.txt
xfail_status SEO02 "sitemap.xml 404 (want 200)"  404 200 GET /sitemap.xml
xfail_status SEO03 "llms.txt 404 (want 200)"     404 200 GET /llms.txt

# The handlers are unreachable at the stripped path too — proving it's the
# registration form, not just the extension, that's wrong.
assert_status SEO04 "robots handler unreachable at /robots"   404 GET /robots
assert_status SEO05 "sitemap handler unreachable at /sitemap" 404 GET /sitemap
assert_status SEO06 "llms handler unreachable at /llms"       404 GET /llms

# Contrast: the ".png" routes are registered WITHOUT the extension, so BOTH the
# extension and stripped forms resolve. This is the exact fix pattern the SEO
# routes need.
assert_status SEO07 "share card resolves at .png"     200 GET /share/card.png
assert_status SEO08 "share card resolves stripped"    200 GET /share/card

# Static assets are unaffected — the FileServer reads the original r.URL.Path.
assert_status SEO09 "css asset unaffected by URLFormat" 200 GET /public/css/main.css
