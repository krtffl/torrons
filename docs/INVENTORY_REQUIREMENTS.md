# 2025 Inventory Requirements - Torrons Vicens

**Document Version:** 1.0
**Last Updated:** November 5, 2025
**Deadline:** End of November 2025
**Campaign Launch:** Early December 2025
**Results Release:** January 6, 2026 (Dia de Reis)

---

## 📋 Required Information for Each Torron

### **Core Information (REQUIRED)**

For each torron product in the 2025 catalog, please provide:

1. **Product Name** (string, max 255 chars)
   - Spanish or Catalan name
   - Include any subtitle/variant (e.g., "- Albert Adrià")
   - Example: `"Mandarina i yuzu - Albert Adrià"`

2. **Category Assignment** (one of):
   - `Clàssics` (ID: 1) - Traditional torrones
   - `Novetats` (ID: 2) - New/revolutionary flavors
   - `Xocolata` (ID: 3) - Chocolate-focused
   - `Albert Adrià` (ID: 4) - Premium designer collection
   - *(Future: Global - will mix all categories)*

3. **Product Image** (JPG/PNG)
   - **Filename:** Use descriptive lowercase name with underscores
   - **Example:** `mandarina_yuzu.jpg`, `brownie_nous.jpg`
   - **Dimensions:** Minimum 500px width, maintain aspect ratio
   - **Format:** JPG preferred (smaller file size), PNG if transparency needed
   - **Background:** White or transparent preferred
   - **Quality:** High quality, well-lit product photography
   - **Consistency:** All photos should have similar style/lighting

4. **Product Status**
   - Is this a NEW product for 2025? (Yes/No)
   - Is this DISCONTINUED from previous years? (Yes/No)
   - Is this RETURNING from 2023? (Yes/No)

---

### **Extended Information (OPTIONAL but RECOMMENDED)**

These fields would enhance the product comparison experience:

5. **Product Description** (string, max 500 chars)
   - Brief description of flavor profile
   - Key ingredients or special characteristics
   - Example: `"Una combinació refrescant de mandarina i yuzu amb base de torró suau"`

6. **Allergen Information** (array of strings)
   - Common allergens present
   - Examples: `Ametlles`, `Llet`, `Ou`, `Fruits secs`, `Gluten`, `Soja`
   - This helps users with dietary restrictions

7. **Special Attributes** (boolean flags)
   - `Vegan` (Yes/No)
   - `Sense Gluten` (Gluten-free)
   - `Sense Lactosa` (Lactose-free)
   - `Ecològic` (Organic)
   - `Artesanal` (Handcrafted)

8. **Weight/Size** (string)
   - Product weight in grams
   - Example: `"200g"`, `"300g"`

9. **Price** (numeric, optional)
   - Retail price in EUR
   - Example: `12.50`
   - Note: May not want to display during voting, but useful for analytics

10. **Main Ingredients** (array of strings, top 3-5)
    - Key flavor components
    - Example: `["Mandarina", "Yuzu", "Mel", "Ametlla"]`

11. **Intensity Level** (1-5 scale, optional)
    - Flavor intensity rating
    - 1 = Subtle/Mild, 5 = Intense/Bold
    - Helps with pairing recommendations

12. **Product URL** (string, optional)
    - Link to product page on Vicens website
    - Example: `"https://www.vicens.com/productos/mandarina-yuzu"`

---

## 🗂️ Data Collection Format

### **Preferred Format: CSV/Excel**

Please provide a spreadsheet with these columns:

```csv
Name,Category,ImageFilename,IsNew2025,IsDiscontinued,Description,Allergens,Vegan,GlutenFree,LactoseFree,Weight,Price,MainIngredients,IntensityLevel,ProductURL
"Mandarina i yuzu - Albert Adrià","Albert Adrià","mandarina_yuzu.jpg","No","No","Combinació refrescant de mandarina i yuzu","Ametlles;Llet","No","No","No","200g","14.50","Mandarina;Yuzu;Mel;Ametlla","3","https://www.vicens.com/productos/mandarina-yuzu"
```

### **Alternative: JSON Format**

```json
{
  "torrons": [
    {
      "name": "Mandarina i yuzu - Albert Adrià",
      "category": "Albert Adrià",
      "imageFilename": "mandarina_yuzu.jpg",
      "isNew2025": false,
      "isDiscontinued": false,
      "description": "Combinació refrescant de mandarina i yuzu",
      "allergens": ["Ametlles", "Llet"],
      "attributes": {
        "vegan": false,
        "glutenFree": false,
        "lactoseFree": false,
        "organic": false
      },
      "weight": "200g",
      "price": 14.50,
      "mainIngredients": ["Mandarina", "Yuzu", "Mel", "Ametlla"],
      "intensityLevel": 3,
      "productUrl": "https://www.vicens.com/productos/mandarina-yuzu"
    }
  ]
}
```

---

## 📸 Image Collection Checklist

### **What to Gather:**

1. **All Product Images**
   - One image per torron
   - Consistent style across all photos
   - Professional quality

2. **Image Naming Convention:**
   ```
   Good Examples:
   - mandarina_yuzu.jpg
   - brownie_nous.jpg
   - xocolata_negra.jpg
   - torrone_clasico.jpg

   Bad Examples:
   - IMG_1234.jpg
   - photo1.jpg
   - MandarinaYuzu.JPG (use lowercase)
   ```

3. **Image Organization:**
   - Place all images in a single folder
   - No subfolders
   - Verify no duplicate filenames
   - Check all images open correctly

---

## 🎨 Brand Guidelines to Extract

When visiting the Torrons Vicens website, please document:

### **Visual Design:**

1. **Color Palette**
   - Primary brand color (hex code)
   - Secondary colors
   - Accent colors for buttons/CTAs
   - Background colors
   - Text colors (primary, secondary)
   - Screenshot the homepage/product pages

2. **Typography**
   - Primary font family
   - Heading fonts
   - Body text fonts
   - Font weights used
   - Font sizes for different elements

3. **UI Patterns**
   - Button styles (shape, size, hover effects)
   - Card layouts for products
   - Spacing/padding patterns
   - Border radius (rounded corners)
   - Shadow effects
   - Navigation style

4. **Product Display**
   - How are products shown in listings?
   - What information is visible on product cards?
   - Image aspect ratios
   - Badge/label styles (NEW, ORGANIC, etc.)

5. **Overall Aesthetic**
   - Modern or traditional?
   - Minimalist or ornate?
   - Color-heavy or neutral?
   - Image-focused or text-focused?

### **Content Elements:**

6. **Product Pages**
   - What details are shown for each product?
   - How is allergen info displayed?
   - How are ingredients listed?
   - Any special sections (pairing suggestions, etc.)?

7. **Voice & Tone**
   - Formal or casual?
   - Language preference (Catalan, Spanish, both?)
   - How do they describe products?

---

## 🔄 Migration Strategy

### **Step 1: Audit Current Database**

Current inventory from 2023 includes approximately:
- **Albert Adrià:** 34 products
- **Novetats:** ~15 products
- **Clàssics:** ~25 products
- **Xocolata:** ~20 products
- **Total:** ~94 products

### **Step 2: Categorize Changes**

Based on 2025 catalog, identify:

1. **Products to KEEP** (2023 → 2025)
   - These will maintain their historical ELO ratings
   - Update any changed information (price, description)
   - Keep existing images or update if better quality available

2. **Products to REMOVE** (Discontinued)
   - Mark as discontinued in database
   - Do NOT delete (preserve historical data)
   - Exclude from 2025 campaign pairings

3. **Products to ADD** (New for 2025)
   - Add with initial ELO rating of 1500
   - Upload new images
   - Generate new pairings

### **Step 3: Data Validation**

Before import:
- [ ] Verify all product names are unique
- [ ] Check all image files exist
- [ ] Validate category assignments
- [ ] Confirm no missing required fields
- [ ] Test image loading in browser

---

## 📊 Proposed Schema Enhancements

Based on the extended information, I recommend adding these fields to the database:

### **New Columns for "Torrons" Table:**

```sql
ALTER TABLE "Torrons"
    ADD COLUMN "Description" text,
    ADD COLUMN "Allergens" text[],  -- Array of allergen strings
    ADD COLUMN "IsVegan" boolean DEFAULT false,
    ADD COLUMN "IsGlutenFree" boolean DEFAULT false,
    ADD COLUMN "IsLactoseFree" boolean DEFAULT false,
    ADD COLUMN "IsOrganic" boolean DEFAULT false,
    ADD COLUMN "Weight" varchar(50),
    ADD COLUMN "Price" numeric(10,2),
    ADD COLUMN "MainIngredients" text[],  -- Array of ingredient strings
    ADD COLUMN "IntensityLevel" int CHECK (IntensityLevel >= 1 AND IntensityLevel <= 5),
    ADD COLUMN "ProductUrl" varchar(500),
    ADD COLUMN "IsNew2025" boolean DEFAULT false,
    ADD COLUMN "Discontinued" boolean DEFAULT false,
    ADD COLUMN "YearAdded" int DEFAULT 2025;
```

### **Benefits:**

1. **Enhanced Filtering**
   - Filter by allergens (show only gluten-free)
   - Filter by dietary preferences (vegan, lactose-free)
   - Filter by price range

2. **Better Product Cards**
   - Display description during voting
   - Show dietary badges (🌱 Vegan, 🌾 Gluten-Free)
   - Link to product page for more info

3. **Richer Results Page**
   - Show winning product details
   - Display full information for top torrons
   - Help users make purchasing decisions

4. **Analytics Potential**
   - Track popularity by dietary preference
   - Analyze price vs. rating correlation
   - Identify trending flavors/ingredients

---

## ⏰ Timeline for Data Collection

**CRITICAL DATES:**

- **November 12:** Inventory data collection deadline
- **November 15:** Data validation and import complete
- **November 18:** Begin user testing with real 2025 data
- **November 25:** Final adjustments and polish
- **November 30:** Production deployment
- **Early December:** Public launch
- **January 6, 2026:** Results reveal

**Please prioritize getting this information by November 12 to stay on schedule!**

---

## 🤝 How to Share the Data

Once you have the information, you can:

1. **Create a Google Sheet** and share the link
2. **Upload a CSV/Excel file** to a shared drive
3. **Create a JSON file** and commit to a branch
4. **Use a GitHub issue** with the data in a comment
5. **Email the spreadsheet** directly

---

## 📝 Quick Reference Checklist

For each torron in the 2025 catalog:

- [ ] Product name in Catalan/Spanish
- [ ] Category (Clàssics, Novetats, Xocolata, or Albert Adrià)
- [ ] Image file (high quality, consistent style)
- [ ] Status (New/Returning/Discontinued)
- [ ] Description (optional but recommended)
- [ ] Allergen information (optional but recommended)
- [ ] Dietary attributes (optional but recommended)

**Absolute minimum needed:** Name, Category, Image, Status

---

## ❓ Questions?

If you have questions while collecting this data:
- Uncertain about category assignment? → Document your best guess and flag for review
- Missing an image? → Use placeholder, we can add later
- Unclear on allergens? → Leave blank, we'll research
- Product in multiple categories? → Pick the PRIMARY category

**The goal is progress, not perfection!** We can refine as we go.
