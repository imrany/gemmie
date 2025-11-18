// Video detection and processing export function
export async function detectAndProcessVideo(url: string, scrapedData: any): Promise<{
  videoUrl?: string,
  type?: 'youtube' | 'vimeo' | 'direct' | 'twitter' | 'tiktok',
  thumbnail?: string,
  duration?: string,
  embedHtml?: string
}> {
  const domain = new URL(url).hostname.toLowerCase()

  // YouTube detection
  if (domain.includes('youtube.com') || domain.includes('youtu.be')) {
    const videoId = extractYouTubeId(url)
    if (videoId) {
      return {
        videoUrl: url,
        type: 'youtube',
        thumbnail: `https://img.youtube.com/vi/${videoId}/maxresdefault.jpg`,
        embedHtml: `<iframe width="100%" height="100%" src="https://www.youtube-nocookie.com/embed/${videoId}" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>`
      }
    }
  }

  // Vimeo detection
  if (domain.includes('vimeo.com')) {
    const videoId = extractVimeoId(url)
    if (videoId) {
      try {
        const vimeoData = await fetch(`https://vimeo.com/api/v2/video/${videoId}.json`)
        const data = await vimeoData.json()
        return {
          videoUrl: url,
          type: 'vimeo',
          thumbnail: data[0]?.thumbnail_large,
          duration: data[0]?.duration ? formatDuration(data[0].duration) : undefined,
          embedHtml: `<iframe src="https://player.vimeo.com/video/${videoId}" width="100%" height="100%" frameborder="0" allow="autoplay; fullscreen; picture-in-picture" allowfullscreen></iframe>`
        }
      } catch (error) {
        console.error('Failed to fetch Vimeo data:', error)
        // Fallback without duration
        return {
          videoUrl: url,
          type: 'vimeo',
          thumbnail: scrapedData.preview_image || scrapedData.images?.[0],
          embedHtml: `<iframe src="https://player.vimeo.com/video/${videoId}" width="100%" height="100%" frameborder="0" allow="autoplay; fullscreen; picture-in-picture" allowfullscreen></iframe>`
        }
      }
    }
  }

  // Twitter/X video detection
  if (domain.includes('twitter.com') || domain.includes('x.com')) {
    return {
      videoUrl: url,
      type: 'twitter',
      thumbnail: scrapedData.preview_image || scrapedData.images?.[0]
    }
  }

  // TikTok detection
  if (domain.includes('tiktok.com')) {
    return {
      videoUrl: url,
      type: 'tiktok',
      thumbnail: scrapedData.preview_image || scrapedData.images?.[0]
    }
  }

  // Direct video file detection
  if (scrapedData.video || url.match(/\.(mp4|webm|ogg|mov|avi)$/i)) {
    return {
      videoUrl: scrapedData.video || url,
      type: 'direct',
      thumbnail: scrapedData.preview_image || scrapedData.images?.[0]
    }
  }

  return {}
}

// Helper export functions
function extractYouTubeId(url: string): string | null {
  const regex = /(?:youtube\.com\/(?:[^\/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?\/\s]{11})/
  const match = url.match(regex)
  return match ? match[1] : null
}

function extractVimeoId(url: string): string | null {
  const regex = /vimeo\.com\/(?:.*#|.*\/videos\/)?([0-9]+)/
  const match = url.match(regex)
  return match ? match[1] : null
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const remainingSeconds = seconds % 60
  
  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, '0')}:${remainingSeconds.toString().padStart(2, '0')}`
  }
  return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`
}

// Global observer instance
let videoLazyObserver: IntersectionObserver | null = null

// Enhanced video lazy loading that doesn't auto-load (click to play)
export function initializeVideoLazyLoading() {
  if (videoLazyObserver) return // Already initialized
  
  videoLazyObserver = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const container = entry.target as HTMLElement
        // Mark as ready for interaction but don't auto-load
        container.classList.add('video-ready')
        // We still observe but don't auto-replace - user must click
      }
    })
  }, {
    rootMargin: '50px'
  })

  // Observe existing video containers
  observeExistingVideoContainers()
}

// Observe existing video containers
export function observeExistingVideoContainers() {
  if (!videoLazyObserver) return
  
  document.querySelectorAll('.video-embed-container:not([data-observed])').forEach(container => {
    videoLazyObserver?.observe(container)
    container.setAttribute('data-observed', 'true')
  })
}

// Call this when adding new link previews dynamically
export function observeNewVideoContainers(parentElement?: HTMLElement) {
  if (!videoLazyObserver) {
    initializeVideoLazyLoading()
    return
  }
  
  const scope = parentElement || document
  scope.querySelectorAll('.video-embed-container:not([data-observed])').forEach(container => {
    videoLazyObserver!.observe(container)
    container.setAttribute('data-observed', 'true')
  })
}

// Cleanup export function
export function destroyVideoLazyLoading() {
  if (videoLazyObserver) {
    videoLazyObserver.disconnect()
    videoLazyObserver = null
  }
  
  // Clear any remaining video ready states
  document.querySelectorAll('.video-ready').forEach(container => {
    container.classList.remove('video-ready')
    container.removeAttribute('data-observed')
  })
}

// Enhanced video control export functions
const videoStates = new Map<string, {
  iframe?: HTMLIFrameElement,
  isPlaying: boolean,
  originalThumbnail?: string
}>()

export function playEmbeddedVideo(thumbnailElement: HTMLElement, videoId: string) {
  const container = thumbnailElement.parentElement as HTMLElement
  let embedHtml = container.dataset.embed
  
  if (embedHtml) {
    // Store original thumbnail for restoration
    const originalThumbnail = container.innerHTML
    
    // Restore single quotes and add autoplay
    embedHtml = embedHtml.replace(/&apos;/g, "'")
    
    if (embedHtml.includes('youtube')) {
      embedHtml = embedHtml.replace(/src="([^"]*)"/, 'src="$1?autoplay=1&enablejsapi=1"')
    } else if (embedHtml.includes('vimeo')) {
      embedHtml = embedHtml.replace(/src="([^"]*)"/, 'src="$1?autoplay=1"')
    }
    
    // Replace thumbnail with iframe
    container.innerHTML = embedHtml
    
    // Store state
    const iframe = container.querySelector('iframe') as HTMLIFrameElement
    videoStates.set(videoId, {
      iframe,
      isPlaying: true,
      originalThumbnail
    })
    
    // Show controls
    showVideoControls(videoId)
  }
}

export function pauseVideo(videoId: string) {
  const state = videoStates.get(videoId)
  if (state?.iframe) {
    // For YouTube
    if (state.iframe.src.includes('youtube')) {
      state.iframe.contentWindow?.postMessage('{"event":"command","func":"pauseVideo","args":""}', '*')
    }
    // For Vimeo
    else if (state.iframe.src.includes('vimeo')) {
      state.iframe.contentWindow?.postMessage('{"method":"pause"}', '*')
    }
    
    state.isPlaying = false
    updateVideoControlButtons(videoId, 'paused')
  }
}

export function resumeVideo(videoId: string) {
  const state = videoStates.get(videoId)
  if (state?.iframe) {
    // For YouTube
    if (state.iframe.src.includes('youtube')) {
      state.iframe.contentWindow?.postMessage('{"event":"command","func":"playVideo","args":""}', '*')
    }
    // For Vimeo
    else if (state.iframe.src.includes('vimeo')) {
      state.iframe.contentWindow?.postMessage('{"method":"play"}', '*')
    }
    
    state.isPlaying = true
    updateVideoControlButtons(videoId, 'playing')
  }
}

export function stopVideo(videoId: string) {
  const state = videoStates.get(videoId)
  const videoContainer = document.getElementById(videoId)
  
  if (state && videoContainer) {
    // Restore original thumbnail
    const embedContainer = videoContainer.querySelector('.video-embed-container')
    if (embedContainer && state.originalThumbnail) {
      embedContainer.innerHTML = state.originalThumbnail
    }
    
    // Hide controls
    hideVideoControls(videoId)
    
    // Clear state
    videoStates.delete(videoId)
  }
}

// Direct video controls
export function toggleDirectVideo(videoId: string) {
  const video = document.getElementById(`${videoId}-video`) as HTMLVideoElement
  const button = document.querySelector(`#${videoId}-controls .toggle-btn`) as HTMLButtonElement
  
  if (video) {
    if (video.paused) {
      video.play()
      updateDirectVideoButton(button, 'playing')
    } else {
      video.pause()
      updateDirectVideoButton(button, 'paused')
    }
  }
}

export function stopDirectVideo(videoId: string) {
  const video = document.getElementById(`${videoId}-video`) as HTMLVideoElement
  const button = document.querySelector(`#${videoId}-controls .toggle-btn`) as HTMLButtonElement
  
  if (video) {
    video.pause()
    video.currentTime = 0
    updateDirectVideoButton(button, 'stopped')
    hideVideoControls(videoId)
  }
}

// Control visibility export functions
export function showVideoControls(videoId: string) {
  const controls = document.getElementById(`${videoId}-controls`)
  if (controls) {
    controls.classList.remove('opacity-0')
    controls.classList.add('opacity-100')
  }
}

export function hideVideoControls(videoId: string) {
  const controls = document.getElementById(`${videoId}-controls`)
  if (controls) {
    controls.classList.remove('opacity-100')
    controls.classList.add('opacity-0')
  }
}

export function updateVideoControlButtons(videoId: string, state: 'playing' | 'paused') {
  const controls = document.getElementById(`${videoId}-controls`)
  if (controls) {
    const pauseBtn = controls.querySelector('.pause-btn') as HTMLElement
    const playBtn = controls.querySelector('.play-btn') as HTMLElement
    
    if (state === 'playing') {
      pauseBtn?.classList.remove('hidden')
      playBtn?.classList.add('hidden')
    } else {
      pauseBtn?.classList.add('hidden')
      playBtn?.classList.remove('hidden')
    }
  }
}

export function updateDirectVideoButton(button: HTMLButtonElement, state: 'playing' | 'paused' | 'stopped') {
  const playIcon = button.querySelector('.play-icon')
  const pauseIcon = button.querySelector('.pause-icon')
  
  if (state === 'playing') {
    playIcon?.classList.add('hidden')
    pauseIcon?.classList.remove('hidden')
  } else {
    playIcon?.classList.remove('hidden')
    pauseIcon?.classList.add('hidden')
  }
}

export function updateVideoControls(videoId: string, state: 'paused' | 'ended') {
  showVideoControls(videoId)
  if (state === 'ended') {
    hideVideoControls(videoId)
  }
}

export function playSocialVideo(url: string, platform: string) {
  window.open(url, '_blank', 'noopener,noreferrer')
}