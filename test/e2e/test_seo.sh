#!/usr/bin/env bash
# SEO surface. FIXED (Batch 5): the /robots.txt, /sitemap.xml and /llms.txt
# routes were registered WITH their extension, which middleware.URLFormat strips
# before chi matches — so they were unreachable dead code (404 at every URL).
# They are now registered at the dotless path (/robots, /sitemap, /llms), so the
# real ".ext" URLs resolve via the strip, the same trick /share/card.png uses.
section "SEO — sitemap / robots / llms reachable; og-image present"

# The canonical public URLs a crawler hits now work.
assert_status        SEO01 "robots.txt -> 200"   200 GET /robots.txt
assert_header        SEO01b "robots.txt is text" "content-type" "text/plain" GET /robots.txt
assert_status        SEO02 "sitemap.xml -> 200"  200 GET /sitemap.xml
assert_header        SEO02b "sitemap is xml"     "content-type" "xml" GET /sitemap.xml
assert_status        SEO03 "llms.txt -> 200"     200 GET /llms.txt
assert_body_contains SEO03b "robots points at sitemap" "Sitemap:" GET /robots.txt

# Dotless forms resolve too (that's the registration form that makes the .ext
# URLs work through URLFormat's stripping).
assert_status SEO04 "/robots resolves"   200 GET /robots
assert_status SEO05 "/sitemap resolves"  200 GET /sitemap
assert_status SEO06 "/llms resolves"     200 GET /llms

# The .png image routes keep working at both forms (the pattern the SEO routes
# now mirror).
assert_status SEO07 "share card at .png"  200 GET /share/card.png
assert_status SEO08 "share card stripped" 200 GET /share/card

# The Open Graph / Twitter card image referenced by every page now exists.
assert_status SEO10 "og-image present"    200 GET /public/assets/og-image.jpg
assert_header SEO10b "og-image is jpeg"   "content-type" "image/jpeg" GET /public/assets/og-image.jpg

# Static assets unaffected by URLFormat (FileServer reads the original path).
assert_status SEO09 "css asset unaffected" 200 GET /public/css/main.css
