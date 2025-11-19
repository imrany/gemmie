# MarkdownRenderer - Implementation Guide

## 📋 Overview

This implementation merges your two markdown rendering approaches and adds an
interactive HTML preview feature. Users can click a "Preview" button on HTML
code blocks to see the rendered output in a modal.

## 🎯 Key Features

### ✅ Merged Functionality

- **All markdown features** from both implementations
- **Component-based rendering** (CodeBlock, Image, Table, Callout, Link)
- **Syntax highlighting** with highlight.js
- **HTML preview** for code blocks

### ✅ HTML Preview

- **Automatic detection** of HTML code (html, htm, or contains `<!DOCTYPE>`)
- **Preview button** appears only for HTML code
- **Full-screen modal** with sandboxed iframe
- **Interactive** - fully functional HTML with JS
- **Secure** - uses iframe sandbox attributes
- **Keyboard support** - ESC to close

## 📁 File Structure

```bash
src/
├── components/
│   └── ui/
│       └── markdown/
│           ├── MarkdownRenderer.vue      (Main component - UPDATED)
│           ├── CodeBlock.vue             (Code display - UPDATED)
│           ├── MarkdownImage.vue         (Image rendering)
│           ├── MarkdownTable.vue         (Table rendering)
│           ├── MarkdownCallout.vue       (Callouts/alerts)
│           └── MarkdownLink.vue          (Link rendering)
```

## 🔧 Implementation Steps

### Step 1: Update MarkdownRenderer.vue

Replace your existing `MarkdownRenderer.vue` with the enhanced version that includes:

```vue
// Key additions: - showPreviewModal ref - previewCode ref - isPreviewableCode()
function - openPreview() function - closePreview() function - Preview modal in
template
```

### Step 2: Update CodeBlock.vue

Update your `CodeBlock.vue` to include:

```vue
// Props: interface Props { data: CodeBlockData; isPreviewable?: boolean; // NEW
} // Emits: const emit = defineEmits<{ preview: []; // NEW }>(); // Template
additions: - Preview button (conditional on isPreviewable) - Eye icon from
lucide-vue-next - handlePreview() method
```

### Step 3: Install Dependencies

Make sure you have these installed:

```bash
npm install lucide-vue-next highlight.js marked
# or
yarn add lucide-vue-next highlight.js marked
```

### Step 4: Import in ChatView.vue

In your `ChatView.vue`, use the component as normal:

```vue
<MarkdownRenderer :content="item.response || ''" />
```

The preview functionality will work automatically!

## 🎨 Features Breakdown

### 1. Code Block Detection

```typescript
const isPreviewableCode = (language: string, code: string): boolean => {
  const htmlLanguages = ["html", "htm", "xml"];
  return (
    htmlLanguages.includes(language.toLowerCase()) ||
    code.trim().startsWith("<!DOCTYPE") ||
    code.includes("<html") ||
    code.includes("<body")
  );
};
```

### 2. Modal Preview

```vue
<iframe
  :srcdoc="previewCode"
  class="w-full h-full border-0"
  sandbox="allow-scripts allow-forms allow-modals allow-popups"
  title="HTML Preview"
/>
```

**Security**: The `sandbox` attribute restricts what the iframe can do:

- ✅ `allow-scripts` - JavaScript can run
- ✅ `allow-forms` - Forms work
- ✅ `allow-modals` - Alerts/modals allowed
- ✅ `allow-popups` - Window.open() works
- ❌ No access to parent page
- ❌ No navigation of parent
- ❌ No same-origin bypass

### 3. Keyboard Support

```typescript
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === "Escape" && showPreviewModal.value) {
    closePreview();
  }
};
```

## 🎭 Usage Examples

### Basic HTML Code

```markdown
Here's a button:

\`\`\`html

<!DOCTYPE html>
<html>
<body>
    <button onclick="alert('Hello!')">Click Me</button>
</body>
</html>
\`\`\`
```

**Result**: Preview button appears → Click to see interactive button

### Python Code (No Preview)

```markdown
\`\`\`python
print("Hello World")
\`\`\`
```

**Result**: Only Copy button (no preview for non-HTML)

### Complex Game

```markdown
\`\`\`html

<!DOCTYPE html>
<!-- Full game code here -->

\`\`\`
```

**Result**: Preview button → Full interactive game in modal

## 🔐 Security Considerations

### ✅ Safe

- Iframe sandbox prevents malicious code from accessing parent
- No localStorage/sessionStorage access from iframe
- No cookie access
- Cannot navigate parent window
- Cannot access parent DOM

### ⚠️ Still Executes JavaScript

- Code inside iframe CAN run JavaScript
- Good for demos, games, interactive examples
- User should trust the source of the HTML

### 🛡️ Additional Protection (Optional)

If you want to block certain scripts:

```typescript
const sanitizeHTML = (html: string): string => {
  // Remove potentially dangerous attributes
  return html
    .replace(/on\w+\s*=/gi, "") // Remove inline event handlers
    .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, "");
  // Remove script tags
};
```

## 🎨 Styling

The modal uses your existing Tailwind classes and adapts to dark mode:

```vue
<!-- Light Mode -->
bg-white text-gray-900

<!-- Dark Mode -->
dark:bg-gray-900 dark:text-white
```

## 📱 Responsive Design

- **Desktop**: Full-screen modal (11/12 width, 5/6 height)
- **Mobile**: Adapts to smaller screens
- **Tablet**: Comfortable viewing size

## 🐛 Troubleshooting

### Preview button not showing

- Check if code language is "html", "htm", or contains `<!DOCTYPE>`
- Verify `isPreviewable` prop is being passed

### Modal not closing

- Check if ESC key handler is registered
- Verify `closePreview()` is called on backdrop click

### Code not rendering

- Check iframe `srcdoc` attribute is receiving code
- Verify sandbox attributes are not too restrictive

### Dark mode issues

- Ensure all `dark:` prefixes are present
- Check Tailwind dark mode is configured

## 🚀 Advanced Features

### Custom Preview Size

Modify modal dimensions:

```vue
<div class="w-11/12 h-5/6"></div>
```

### Add Download Button

```vue
<button @click="downloadHTML">
    <Download :size="14" />
    Download
</button>
```

```typescript
const downloadHTML = () => {
  const blob = new Blob([previewCode.value], { type: "text/html" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "preview.html";
  a.click();
};
```

### Add Fullscreen Mode

```typescript
const goFullscreen = () => {
  const elem = document.querySelector(".preview-modal");
  elem?.requestFullscreen();
};
```

## ✨ Benefits

1. **Better UX** - Users see code results immediately
2. **Educational** - Great for learning/teaching
3. **Interactive** - Games and demos work fully
4. **Clean** - Non-intrusive, only shows for HTML
5. **Secure** - Sandboxed execution
6. **Responsive** - Works on all devices
7. **Accessible** - Keyboard navigation (ESC)

## 📊 Performance

- **Lazy rendering** - Preview only created when clicked
- **Isolated** - Each preview is independent
- **Efficient** - Modal reuses same DOM element
- **Memory** - Modal destroyed on close

## 🎓 Best Practices

1. **Always use sandbox** - Never remove sandbox attributes
2. **Test thoroughly** - Try malicious code to verify isolation
3. **User feedback** - Show loading states for complex HTML
4. **Error handling** - Catch iframe errors gracefully
5. **Accessibility** - Provide keyboard shortcuts

## 📝 Summary

This implementation gives you:

- ✅ Complete markdown rendering
- ✅ Syntax highlighting
- ✅ HTML preview with one click
- ✅ Secure sandboxed execution
- ✅ Clean, modern UI
- ✅ Dark mode support
- ✅ Keyboard shortcuts
- ✅ Responsive design

Perfect for your chat application where AI generates HTML code!
