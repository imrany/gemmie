# Modular HomeView Structure Documentation

## Overview

The original `HomeView.vue` component was a monolithic file with over 3,600 lines of code. It has been refactored into a modular structure using Vue 3 composables and smaller, focused components.

## File Structure

```
src/
├── composables/
│   ├── useAuth.ts           # Authentication logic
│   ├── useVoiceRecording.ts # Voice recognition functionality
│   ├── usePasteHandling.ts  # Large paste content handling
│   ├── useChat.ts           # Chat operations and message management
│   └── usePagination.ts     # Deep search result pagination
├── components/chat/
│   ├── ChatInput.vue        # Message input component
│   └── ChatMessage.vue      # Individual message display
└── views/
    ├── HomeView.vue         # Main view (now modular)
    └── HomeView.backup.vue  # Original backup
```

## Composables

### `useAuth.ts`
Handles user authentication flow including:
- Multi-step registration process
- Form validation
- Error handling
- Session management

**Key Functions:**
- `nextAuthStep()` / `prevAuthStep()` - Navigation between auth steps
- `validateCurrentStep()` - Form validation
- `handleStepSubmit()` - Form submission
- `handleAuthSuccess()` / `handleAuthError()` - Response handling

### `useVoiceRecording.ts`
Manages speech recognition and voice input:
- Browser speech recognition API integration
- Recording state management
- Transcription handling
- Permission management

**Key Functions:**
- `initializeSpeechRecognition()` - Setup speech recognition
- `toggleVoiceRecording()` - Start/stop recording
- `clearVoiceTranscription()` - Reset transcription

### `usePasteHandling.ts`
Handles large content paste operations:
- Large paste detection
- Content preview modals
- Draft management
- Content type detection

**Key Functions:**
- `handlePaste()` - Process paste events
- `openPasteModal()` - Show paste preview
- `removePastePreview()` - Clean up previews

### `useChat.ts`
Core chat functionality:
- Message management
- Chat creation and switching
- API communication
- Draft persistence

**Key Functions:**
- `handleSubmit()` - Send messages
- `createNewChat()` / `switchToChat()` - Chat management
- `loadChats()` / `saveChatDrafts()` - Persistence
- `copyResponse()` / `shareResponse()` - Message actions

### `usePagination.ts`
Deep search result navigation:
- Result pagination
- State management per chat
- Navigation controls

**Key Functions:**
- `getPagination()` - Get pagination state
- `nextResult()` / `prevResult()` - Navigate results
- `goToPage()` - Jump to specific page

## Components

### `ChatInput.vue`
Reusable input component featuring:
- Auto-growing textarea
- Voice recording integration
- Input mode selection
- Paste handling
- Loading states

**Props:**
- `disabled` - Input state
- `placeholder` - Placeholder text
- `isLoading` - Loading state
- `currentChatId` - Active chat
- `inputMode` - Response mode

**Events:**
- `@submit` - Message submission
- `@paste` - Paste events

### `ChatMessage.vue`
Individual message display with:
- Markdown rendering
- Deep search pagination
- Message actions (copy, share, refresh)
- Link previews
- Paste content previews

**Props:**
- `message` - Message data
- `messageIndex` - Position in chat
- `currentChatId` - Active chat
- `isLoading` - Loading state

**Events:**
- `@copy` / `@share` / `@refresh` - Message actions
- `@paste-preview-click` - Paste preview interaction

## Benefits of Modular Structure

### 1. **Separation of Concerns**
Each composable handles a specific domain:
- Authentication logic is isolated in `useAuth`
- Voice functionality is contained in `useVoiceRecording`
- Chat operations are centralized in `useChat`

### 2. **Reusability**
Composables can be reused across components:
```vue
// In any component
import { useAuth } from '@/composables/useAuth'
const { isLoading, handleSubmit } = useAuth()
```

### 3. **Testability**
Each composable can be unit tested independently:
```javascript
import { useAuth } from '@/composables/useAuth'

test('validates email correctly', () => {
  const { validateCurrentStep, authData } = useAuth()
  authData.value.email = 'invalid-email'
  expect(validateCurrentStep()).toBe(false)
})
```

### 4. **Maintainability**
- Smaller, focused files are easier to understand
- Changes to auth logic only affect `useAuth.ts`
- Clear boundaries between different functionalities

### 5. **Type Safety**
Each composable has proper TypeScript interfaces:
```typescript
interface AuthData {
  username: string
  email: string
  password: string
  agreeToTerms: boolean
}
```

## Migration Notes

### Breaking Changes
- Global state injection still required for some functionality
- Component prop interfaces may need updates
- Some functions moved between composables

### Backwards Compatibility
- Original `HomeView.backup.vue` preserved
- Same UI/UX experience maintained
- All original functionality preserved

## Usage Example

```vue
<script setup>
import { useAuth, useChat, useVoiceRecording } from '@/composables'
import ChatInput from '@/components/chat/ChatInput.vue'
import ChatMessage from '@/components/chat/ChatMessage.vue'

// Use composables
const { isAuthenticated } = useAuth()
const { currentMessages, handleSubmit } = useChat()
const { initializeSpeechRecognition } = useVoiceRecording()

// Initialize on mount
onMounted(() => {
  initializeSpeechRecognition()
})
</script>

<template>
  <div v-if="isAuthenticated">
    <ChatMessage 
      v-for="message in currentMessages"
      :key="message.id"
      :message="message"
    />
    <ChatInput @submit="handleSubmit" />
  </div>
</template>
```

## Future Improvements

1. **State Management**: Consider Pinia for complex state
2. **Error Boundaries**: Add error handling components
3. **Performance**: Implement virtual scrolling for large chat histories
4. **Testing**: Add comprehensive unit and integration tests
5. **Accessibility**: Enhance keyboard navigation and screen reader support

## Development Guidelines

### Adding New Features
1. Create focused composables for new domains
2. Keep components small and single-purpose
3. Maintain TypeScript interfaces
4. Add proper error handling
5. Write tests for new functionality

### Modifying Existing Features
1. Locate the appropriate composable
2. Update interfaces if needed
3. Test affected components
4. Update documentation

This modular structure provides a solid foundation for future development while maintaining the existing functionality and improving code organization.