#!/bin/bash

# Image Optimization Script for Torrorèndum 2025
# This script optimizes all images in public/images/ directory

set -e

echo "🖼️  Starting image optimization..."
echo ""

# Check for required tools
if ! command -v convert &> /dev/null; then
    echo "❌ ImageMagick not found. Installing..."
    if command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y imagemagick
    elif command -v brew &> /dev/null; then
        brew install imagemagick
    else
        echo "Please install ImageMagick manually: https://imagemagick.org/script/download.php"
        exit 1
    fi
fi

# Navigate to images directory
cd "$(dirname "$0")/public/images"

# Count total images
total_png=$(find . -name "*.png" | wc -l)
total_jpg=$(find . -name "*.jpg" -o -name "*.jpeg" | wc -l)
total=$((total_png + total_jpg))

echo "Found $total images to optimize:"
echo "  - PNG files: $total_png"
echo "  - JPG files: $total_jpg"
echo ""

# Calculate original size
original_size=$(du -sh . | cut -f1)
echo "Original size: $original_size"
echo ""

# Create backup
echo "📦 Creating backup..."
backup_dir="../images_backup_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$backup_dir"
cp -r . "$backup_dir"
echo "✅ Backup created at: $backup_dir"
echo ""

# Optimize PNG files
echo "🔧 Optimizing PNG files..."
count=0
for img in $(find . -name "*.png"); do
    count=$((count + 1))
    echo "  [$count/$total_png] $(basename "$img")"
    convert "$img" -strip -resize 800x800\> -quality 85 "$img"
done
echo ""

# Optimize JPG/JPEG files
echo "🔧 Optimizing JPG files..."
count=0
for img in $(find . -name "*.jpg" -o -name "*.jpeg"); do
    count=$((count + 1))
    echo "  [$count/$total_jpg] $(basename "$img")"
    convert "$img" -strip -resize 800x800\> -quality 85 -sampling-factor 4:2:0 "$img"
done
echo ""

# Calculate new size
new_size=$(du -sh . | cut -f1)
echo "✅ Optimization complete!"
echo ""
echo "Results:"
echo "  - Original size: $original_size"
echo "  - New size: $new_size"
echo "  - Backup location: $backup_dir"
echo ""
echo "Next steps:"
echo "  1. Review the optimized images"
echo "  2. If satisfied: git add public/images && git commit -m 'chore: optimize images'"
echo "  3. If not satisfied: cp -r $backup_dir/* public/images/"
