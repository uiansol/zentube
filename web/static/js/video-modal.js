// Video Player functionality
(function() {
  'use strict';

  let player = null;
  let iframe = null;
  let playerTitle = null;

  // Initialize player on page load
  document.addEventListener('DOMContentLoaded', function() {
    player = document.getElementById('video-player');
    iframe = document.getElementById('video-iframe');
    playerTitle = document.getElementById('video-player-title');

    // Close on Escape key
    document.addEventListener('keydown', function(e) {
      if (e.key === 'Escape' && player && player.classList.contains('active')) {
        closeVideoPlayer();
      }
    });
  });

  // Open video player
  window.openVideoPlayer = function(videoId, videoTitle) {
    if (!player || !iframe) return;
    
    iframe.src = `https://www.youtube.com/embed/${videoId}?autoplay=1`;
    
    if (playerTitle && videoTitle) {
      playerTitle.textContent = videoTitle;
    }
    
    player.classList.add('active');
    
    // Smooth scroll to player
    setTimeout(function() {
      player.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }, 100);
  };

  // Close video player
  window.closeVideoPlayer = function() {
    if (!player || !iframe) return;
    
    player.classList.remove('active');
    iframe.src = '';
    
    if (playerTitle) {
      playerTitle.textContent = '';
    }
  };
})();
