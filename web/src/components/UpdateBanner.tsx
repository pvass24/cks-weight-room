'use client'

import { useState } from 'react'

interface UpdateBannerProps {
  currentVersion: string
  latestVersion: string
  releaseNotes: string
  downloadUrl: string
  isCritical: boolean
  publishedAt: string
  installInstructions: string
  onDismiss?: () => void
}

export default function UpdateBanner({
  currentVersion,
  latestVersion,
  releaseNotes,
  downloadUrl,
  isCritical,
  publishedAt,
  installInstructions,
  onDismiss
}: UpdateBannerProps) {
  const [expanded, setExpanded] = useState(false)

  const bannerColor = isCritical
    ? 'bg-red-600 border-red-700'
    : 'bg-blue-600 border-blue-700'

  const textColor = 'text-white'
  const buttonColor = isCritical
    ? 'bg-red-700 hover:bg-red-800'
    : 'bg-blue-700 hover:bg-blue-800'

  return (
    <div className={`${bannerColor} border-b-2 ${textColor} shadow-lg`}>
      <div className="max-w-7xl mx-auto px-4 py-3">
        {/* Main Banner */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3 flex-1">
            {/* Icon */}
            <div className="flex-shrink-0">
              {isCritical ? (
                <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              ) : (
                <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                </svg>
              )}
            </div>

            {/* Message */}
            <div className="flex-1 min-w-0">
              <p className="font-semibold">
                {isCritical ? 'Critical Update Available' : 'Update Available'}:
                <span className="ml-2 font-mono">{currentVersion} → {latestVersion}</span>
              </p>
              <p className="text-sm opacity-90">
                {isCritical
                  ? 'This is a critical security update. Please update as soon as possible.'
                  : 'A new version is available with improvements and bug fixes.'}
              </p>
            </div>
          </div>

          {/* Actions */}
          <div className="flex items-center gap-2 ml-4">
            <button
              onClick={() => setExpanded(!expanded)}
              className={`px-4 py-2 ${buttonColor} font-semibold rounded transition-colors whitespace-nowrap`}
            >
              {expanded ? 'Hide Details' : 'View Details'}
            </button>
            {onDismiss && !isCritical && (
              <button
                onClick={onDismiss}
                className="p-2 hover:bg-white hover:bg-opacity-20 rounded transition-colors"
                title="Dismiss"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            )}
          </div>
        </div>

        {/* Expanded Details */}
        {expanded && (
          <div className="mt-4 pt-4 border-t border-white border-opacity-30">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              {/* Left Column - Release Notes */}
              <div>
                <h3 className="font-semibold mb-2 flex items-center gap-2">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  Release Notes
                </h3>
                <div className="bg-white bg-opacity-10 rounded p-3 max-h-64 overflow-y-auto">
                  <pre className="text-sm whitespace-pre-wrap font-sans">{releaseNotes || 'No release notes available'}</pre>
                </div>
                {publishedAt && (
                  <p className="text-xs opacity-75 mt-2">
                    Published: {publishedAt}
                  </p>
                )}
              </div>

              {/* Right Column - Installation */}
              <div>
                <h3 className="font-semibold mb-2 flex items-center gap-2">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                  </svg>
                  How to Update
                </h3>
                <div className="bg-white bg-opacity-10 rounded p-3">
                  <pre className="text-sm whitespace-pre-wrap font-sans">{installInstructions}</pre>
                </div>

                {downloadUrl && (
                  <a
                    href={downloadUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="mt-3 inline-flex items-center gap-2 px-4 py-2 bg-white text-blue-600 font-semibold rounded hover:bg-opacity-90 transition-colors"
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    Download {latestVersion}
                  </a>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
