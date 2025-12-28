'use client'

import { useState } from 'react'

interface ValidationBannerProps {
  inGracePeriod: boolean
  graceDaysLeft?: number
  needsValidation: boolean
  onValidate: () => void
  isValidating: boolean
}

export default function ValidationBanner({
  inGracePeriod,
  graceDaysLeft,
  needsValidation,
  onValidate,
  isValidating
}: ValidationBannerProps) {
  const [dismissed, setDismissed] = useState(false)

  if (dismissed || (!inGracePeriod && !needsValidation)) {
    return null
  }

  return (
    <div className={`${
      inGracePeriod ? 'bg-yellow-50 border-yellow-200' : 'bg-blue-50 border-blue-200'
    } border-t border-b px-4 py-3 mb-4`}>
      <div className="max-w-7xl mx-auto flex items-center justify-between">
        <div className="flex items-center gap-3">
          <svg className={`h-5 w-5 ${inGracePeriod ? 'text-yellow-600' : 'text-blue-600'}`} fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
          </svg>
          <div>
            {inGracePeriod ? (
              <p className="text-sm text-yellow-800">
                <span className="font-semibold">License validation pending.</span> Days remaining: {graceDaysLeft}. Connect to the internet to validate your license.
              </p>
            ) : (
              <p className="text-sm text-blue-800">
                <span className="font-semibold">License validation recommended.</span> It's been more than 7 days since your last validation.
              </p>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={onValidate}
            disabled={isValidating}
            className={`px-4 py-2 text-sm font-medium rounded ${
              inGracePeriod
                ? 'bg-yellow-600 hover:bg-yellow-700 text-white'
                : 'bg-blue-600 hover:bg-blue-700 text-white'
            } disabled:opacity-50 disabled:cursor-not-allowed`}
          >
            {isValidating ? 'Validating...' : 'Validate Now'}
          </button>
          <button
            onClick={() => setDismissed(true)}
            className="text-gray-500 hover:text-gray-700"
          >
            <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  )
}
