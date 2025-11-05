# Image Optimization Guide

## Current Status

The application has **20 images larger than 200KB**, with some exceeding 1MB. These should be optimized before production deployment for better performance.

## Large Images Requiring Optimization

The following images are particularly large and should be optimized:

- `xoco_festuc.png` - 1010K → Target: ~150K
- `iogurt.png` - 974K → Target: ~150K
- `tou.jpg` - 985K → Target: ~150K
- `brutal.png` - 967K → Target: ~150K
- `xoco_cruixent.png` - 950K → Target: ~150K
- `llet_cruixent.png` - 914K → Target: ~150K
- `chupa_chups.jpg` - 897K → Target: ~150K
- `iogurt_coco.png` - 859K → Target: ~150K
- `xoco_festuc_3.png` - 858K → Target: ~150K
- `iogurt_festuc.png` - 857K → Target: ~150K
- `xoco_llet_2.png` - 843K → Target: ~150K
- `xoco_bitter.png` - 839K → Target: ~150K
- `tou_cruixent.jpg` - 816K → Target: ~150K
- `xoco_llet.png` - 780K → Target: ~150K
- `blanca_cruixent.png` - 768K → Target: ~150K
- `cremada_taronja.png` - 866K → Target: ~150K

And several more between 200-500KB.

## Recommended Optimization Steps

### Option 1: Automated Optimization (Recommended)

Use a batch image optimizer like ImageOptim, Squoosh CLI, or Sharp:

```bash
# Using Sharp CLI (Node.js)
npm install -g sharp-cli
find public/images -name "*.png" -exec sharp -i {} -o {} -f png -c \;
find public/images -name "*.jpg" -exec sharp -i {} -o {} -f jpeg -q 85 \;

# Using ImageMagick
find public/images -name "*.png" -exec convert {} -strip -resize 800x800\> -quality 85 {} \;
find public/images -name "*.jpg" -exec convert {} -strip -resize 800x800\> -quality 85 {} \;
```

### Option 2: Manual Optimization

1. **Resize images** to appropriate display size (max 800x800px for torron cards)
2. **Compress images**:
   - JPEG: 85% quality
   - PNG: Use pngquant or similar
3. **Convert to WebP** for better compression (with JPEG/PNG fallbacks)
4. **Remove metadata** (EXIF data, color profiles)

### Option 3: Use CDN with Automatic Optimization

Services like Cloudflare Images, Cloudinary, or imgix can automatically optimize images on-the-fly.

## Implementation with WebP

For best performance, serve WebP images with fallbacks:

```html
<picture>
  <source srcset="/public/images/brutal.webp" type="image/webp">
  <source srcset="/public/images/brutal.jpg" type="image/jpeg">
  <img src="/public/images/brutal.jpg" alt="Brutal">
</picture>
```

## Expected Performance Gains

- **Current total image size**: ~19MB for all images
- **After optimization**: ~4-5MB (75% reduction)
- **Faster page loads**: 2-3x improvement on mobile
- **Better user experience**: Especially on slow connections

## Priority

**HIGH** - These optimizations should be completed before production launch to ensure good mobile performance and reduce bandwidth costs.

## Next Steps

1. Run batch optimization on all images
2. Convert large PNGs to JPEGs where appropriate (photos vs graphics)
3. Consider implementing lazy loading for images below the fold
4. Add WebP support for modern browsers

Last updated: 2025-11-05
