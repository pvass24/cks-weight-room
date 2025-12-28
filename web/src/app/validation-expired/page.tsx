'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'

export default function ValidationExpiredPage() {
  const router = useRouter()
  const [isValidating, setIsValidating] = useState(false)
  const [error, setError] = useState('')

  const handleValidateNow = async () => {
    setIsValidating(true)
    setError('')

    try {
      const response = await fetch('/api/activation/validate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' }
      })

      const data = await response.json()

      if (data.success) {
        // Validation succeeded, redirect to home
        router.push('/')
      } else {
        setError(data.message || data.error || 'Validation failed')
      }
    } catch (err) {
      setError('Network error. Please check your internet connection.')
    } finally {
      setIsValidating(false)
    }
  }

  const handleOfflineActivation = () => {
    router.push('/activate?mode=offline')
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-md p-8">
        <div className="text-center mb-6">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-red-100 rounded-full mb-4">
            <svg className="h-8 w-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">License Validation Expired</h1>
          <p className="text-gray-600">
            30 days have passed since your last successful validation. Please connect to the internet to validate your license.
          </p>
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}

        <div className="space-y-3">
          <button
            onClick={handleValidateNow}
            disabled={isValidating}
            className="w-full py-3 px-4 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-semibold rounded-lg transition-colors disabled:cursor-not-allowed"
          >
            {isValidating ? (
              <span className="flex items-center justify-center gap-2">
                <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Validating...
              </span>
            ) : (
              'Validate Now (Online)'
            )}
          </button>

          <button
            onClick={handleOfflineActivation}
            disabled={isValidating}
            className="w-full py-3 px-4 bg-gray-600 hover:bg-gray-700 disabled:bg-gray-400 text-white font-semibold rounded-lg transition-colors disabled:cursor-not-allowed"
          >
            Offline Re-activation
          </button>
        </div>

        <div className="mt-6 p-4 bg-gray-50 rounded-lg">
          <h3 className="text-sm font-semibold text-gray-900 mb-2">Why do I need to validate?</h3>
          <p className="text-xs text-gray-600">
            License validation ensures your subscription is active and helps prevent unauthorized use. You can practice offline for up to 30 days between validations.
          </p>
        </div>
      </div>
    </div>
  )
}
