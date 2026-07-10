#!/usr/bin/env bash
# Compression + cache-header regression checks (Batch 6). These use raw curl so
# they can send Accept-Encoding and read response headers directly.
section "CACHING — compression + Cache-Control"

_enc() { # path -> Content-Encoding for a gzip-accepting request
	curl -s -o /dev/null -D - -H "$(_xff)" -H "Accept-Encoding: gzip" "${BASE_URL}$1" \
		| grep -i '^content-encoding:' | tr -d '\r' | sed 's/^[^:]*: //I'
}
_cc() { # path -> Cache-Control
	curl -s -o /dev/null -D - -H "$(_xff)" "${BASE_URL}$1" \
		| grep -i '^cache-control:' | tr -d '\r' | sed 's/^[^:]*: //I'
}
_assert_gzip() { # id desc path
	local e; e="$(_enc "$2")"
	if grep -qi gzip <<<"$e"; then _ok "[$1] $3 gzipped"; else _no "[$1] $3 NOT gzipped (Content-Encoding: '${e:-none}')"; fi
}
_assert_not_gzip() { # id desc path
	local e; e="$(_enc "$2")"
	if grep -qi gzip <<<"$e"; then _no "[$1] $3 should NOT be gzipped (already compressed)"; else _ok "[$1] $3 not gzipped (correct)"; fi
}
_assert_cc() { # id path needle
	local c; c="$(_cc "$2")"
	if grep -qi "$3" <<<"$c"; then _ok "[$1] $2 Cache-Control has '$3' ($c)"; else _no "[$1] $2 Cache-Control '$c' missing '$3'"; fi
}

# Text responses compress
_assert_gzip CACHE01 /                    "home HTML"
_assert_gzip CACHE02 /public/css/main.css "main.css"
_assert_gzip CACHE03 /sitemap.xml         "sitemap XML"
_assert_gzip CACHE04 /api/campaign/countdown "campaign JSON"

# Already-compressed binaries are left alone
_assert_not_gzip CACHE05 /share/card.png            "share card PNG"
_assert_not_gzip CACHE06 /public/assets/og-image.jpg "og-image JPEG"

# Static assets carry a Cache-Control (were served with none)
_assert_cc CACHE07 /public/css/main.css        "max-age"
_assert_cc CACHE08 /public/assets/og-image.jpg "max-age=2592000"
