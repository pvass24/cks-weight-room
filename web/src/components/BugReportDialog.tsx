'use client'

import { useState } from 'react'

interface BugReportDialogProps {
  isOpen: boolean
  onClose: () => void
  prefilledDescription?: string
}

export default function BugReportDialog({
  isOpen,
  onClose,
  prefilledDescription = ''
}: BugReportDialogProps) {
  const [description, setDescription] = useState(prefilledDescription)
  const [expectedBehavior, setExpectedBehavior] = useState('')
  const [actualBehavior, setActualBehavior] = useState('')
  const [stepsToReproduce, setStepsToReproduce] = useState('')
  const [email, setEmail] = useState('')
  const [includeLogs, setIncludeLogs] = useState(true)
  const [includeDbStats, setIncludeDbStats] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [result, setResult] = useState<{ success: boolean; message: string; filePath?: string } | null>(null)

  const handleSubmit = async () => {
    if (!description.trim()) {
      alert('Please provide a description of the issue')
      return
    }

    setSubmitting(true)
    setResult(null)

    try {
      const response = await fetch('/api/bugreport/submit', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          description,
          expectedBehavior,
          actualBehavior,
          stepsToReproduce,
          email,
          includeLogs,
          includeDbStats,
        }),
      })

      const data = await response.json()

      if (data.success) {
        setResult({
          success: true,
          message: data.message,
          filePath: data.filePath,
        })
      } else {
        setResult({
          success: false,
          message: data.error || 'Failed to generate bug report',
        })
      }
    } catch (err) {
      setResult({
        success: false,
        message: err instanceof Error ? err.message : 'Network error',
      })
    } finally {
      setSubmitting(false)
    }
  }

  const handleClose = () => {
    setDescription(prefilledDescription)
    setExpectedBehavior('')
    setActualBehavior('')
    setStepsToReproduce('')
    setEmail('')
    setResult(null)
    onClose()
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-2xl font-bold text-gray-900">Submit Bug Report</h2>
            <button
              onClick={handleClose}
              className="text-gray-400 hover:text-gray-600"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Success/Error Message */}
          {result && (
            <div className={`mb-6 p-4 rounded-lg ${result.success ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'}`}>
              <p className={`font-semibold ${result.success ? 'text-green-900' : 'text-red-900'}`}>
                {result.success ? 'Success!' : 'Error'}
              </p>
              <p className={`text-sm mt-1 ${result.success ? 'text-green-800' : 'text-red-800'}`}>
                {result.message}
              </p>
              {result.filePath && (
                <p className="text-xs text-green-700 mt-2 font-mono break-all">
                  {result.filePath}
                </p>
              )}
              {result.success && (
                <button
                  onClick={handleClose}
                  className="mt-3 px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-semibold rounded transition-colors"
                >
                  Close
                </button>
              )}
            </div>
          )}

          {!result && (
            <>
              {/* Description */}
              <div className="mb-4">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  What happened? <span className="text-red-500">*</span>
                </label>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  rows={4}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Describe the problem you encountered..."
                  required
                />
              </div>

              {/* Expected Behavior */}
              <div className="mb-4">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  What did you expect to happen?
                </label>
                <textarea
                  value={expectedBehavior}
                  onChange={(e) => setExpectedBehavior(e.target.value)}
                  rows={2}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="What should have happened instead..."
                />
              </div>

              {/* Actual Behavior */}
              <div className="mb-4">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  What actually happened?
                </label>
                <textarea
                  value={actualBehavior}
                  onChange={(e) => setActualBehavior(e.target.value)}
                  rows={2}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Describe the actual behavior..."
                />
              </div>

              {/* Steps to Reproduce */}
              <div className="mb-4">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  Steps to reproduce
                </label>
                <textarea
                  value={stepsToReproduce}
                  onChange={(e) => setStepsToReproduce(e.target.value)}
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="1. Go to...&#10;2. Click on...&#10;3. See error..."
                />
              </div>

              {/* Email (Optional) */}
              <div className="mb-4">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  Email (optional)
                </label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="your.email@example.com"
                />
                <p className="text-xs text-gray-500 mt-1">
                  We'll only use this to follow up on your bug report
                </p>
              </div>

              {/* Options */}
              <div className="mb-6 space-y-2">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={includeLogs}
                    onChange={(e) => setIncludeLogs(e.target.checked)}
                    className="mr-2"
                  />
                  <span className="text-sm text-gray-700">Include log files (last 1000 lines)</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={includeDbStats}
                    onChange={(e) => setIncludeDbStats(e.target.checked)}
                    className="mr-2"
                  />
                  <span className="text-sm text-gray-700">Include database statistics</span>
                </label>
              </div>

              {/* Info Box */}
              <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
                <p className="text-sm text-blue-800">
                  <strong>What will be collected:</strong> System information (OS, architecture, Docker/KIND status),
                  recent log entries, and database statistics (if selected). No sensitive data like passwords or
                  license keys will be included.
                </p>
              </div>

              {/* Actions */}
              <div className="flex gap-3 justify-end">
                <button
                  onClick={handleClose}
                  className="px-4 py-2 bg-gray-200 hover:bg-gray-300 text-gray-800 font-semibold rounded-lg transition-colors"
                  disabled={submitting}
                >
                  Cancel
                </button>
                <button
                  onClick={handleSubmit}
                  disabled={submitting || !description.trim()}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-semibold rounded-lg transition-colors flex items-center gap-2"
                >
                  {submitting ? (
                    <>
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                      Generating Report...
                    </>
                  ) : (
                    'Generate Bug Report'
                  )}
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
