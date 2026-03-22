import { useEffect } from 'react';
import { useAuth } from '../context/AuthContext';

// Dynamic CSS Watermark overlay to deter mobile screenshots
export const Watermark = () => {
  const { user } = useAuth();
  
  useEffect(() => {
    if (!user) return;
    
    // Generate an SVG watermark dynamically with doctor info
    const username = user.username || 'unknown';
    // Note: React can't easily get the real IP natively without an API call, 
    // so we use a placeholder or could fetch it from our own auth endpoint
    const ip = '192.168.x.x'; 
    const timestamp = new Date().toISOString();
    
    // Create an SVG-based pattern
    const svgText = `
      <svg xmlns="http://www.w3.org/2000/svg" width="300" height="300">
        <text x="50%" y="50%" transform="rotate(-45 150 150)" width="300" 
              font-family="system-ui, sans-serif" font-size="14" font-weight="bold" 
              fill="rgba(0,0,0,1)" text-anchor="middle" dominant-baseline="middle">
          ${username} | ${ip}
        </text>
        <text x="50%" y="58%" transform="rotate(-45 150 150)" width="300" 
              font-family="monospace" font-size="10" 
              fill="rgba(0,0,0,1)" text-anchor="middle" dominant-baseline="middle">
          ${timestamp}
        </text>
      </svg>
    `;
    
    // Escape the SVG for CSS usage
    const encodedSvg = `url("data:image/svg+xml;base64,${btoa(svgText)}")`;
    document.documentElement.style.setProperty('--watermark-svg', encodedSvg);
    
  }, [user]);

  if (!user) return null;

  // The CSS class .watermark-overlay creates the tiled background
  return <div className="watermark-overlay" aria-hidden="true" />;
};

// AntiLeak wrapper component for sensitive patient data
export const AntiLeak = ({ children, className = '' }) => {
  return (
    <div 
      className={`secure-text ${className}`}
      onContextMenu={(e) => {
        // Disable right-click completely
        e.preventDefault();
      }}
      onCopy={(e) => {
        // Block copying
        e.preventDefault();
        e.clipboardData.setData('text/plain', 'Confidential Medical Record - Unauthorized copying blocked.');
      }}
      onDragStart={(e) => {
        // Block dragging elements (images, text selections)
        e.preventDefault();
      }}
    >
      {children}
    </div>
  );
};
