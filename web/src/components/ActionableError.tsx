'use client'

import { useState } from 'react'
import BugReportDialog from './BugReportDialog'

interface ActionableErrorProps {
  code: string
  what: string
  why: string
  howToFix: string[]
  retryable?: boolean
  context?: Record<string, string>
  onRetry?: () => void
  onDismiss?: () => void
  additionalActions?: Array<{
    label: string
    onClick: () => void
    variant?: 'primary' | 'secondary'
  }>
  enableBugReport?: boolean
}

export default function ActionableError({
  code,
  what,
  why,
  howToFix,
  retryable = false,
  context,
  onRetry,
  onDismiss,
  additionalActions = [],
  enableBugReport = true
}: ActionableErrorProps) {
  const [showBugReport, setShowBugReport] = useState(false)

  const bugReportDescription = `Error Code: ${code}\n\nWhat: ${what}\n\nWhy: ${why}${
    context ? '\n\nContext:\n' + Object.entries(context).map(([k, v]) => `- ${k}: ${v}`).join('\n') : ''
  }`

  return (
    <>
      <div className="bg-red-50 border border-red-200 rounded-lg p-6 shadow-md">
      <div className="flex items-start gap-3 mb-4">
        <div className="flex-shrink-0">
          <svg className="h-6 w-6 text-red-600" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 9.586 8.707 8.293z" clipRule="evenodd" />
          </svg>
        </div>
        <div className="flex-1">
          <h3 className="text-lg font-semibold text-red-900 mb-1">
            Error ({code})
          </h3>
        </div>
        {onDismiss && (
          <button
            onClick={onDismiss}
            className="flex-shrink-0 text-red-400 hover:text-red-600"
          >
            <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
            </svg>
          </button>
        )}
      </div>

      <div className="space-y-4">
        {/* What went wrong */}
        <div>
          <p className="text-sm font-semibold text-red-900 mb-1">What:</p>
          <p className="text-sm text-red-800">{what}</p>
        </div>

        {/* Why it happened */}
        <div>
          <p className="text-sm font-semibold text-red-900 mb-1">Why:</p>
          <p className="text-sm text-red-800">{why}</p>
        </div>

        {/* Context information */}
        {context && Object.keys(context).length > 0 && (
          <div className="bg-red-100 rounded p-3">
            {Object.entries(context).map(([key, value]) => (
              <div key={key} className="text-sm">
                <span className="font-semibold text-red-900 capitalize">{key}:</span>{' '}
                <span className="text-red-800">{value}</span>
              </div>
            ))}
          </div>
        )}

        {/* How to fix */}
        <div>
          <p className="text-sm font-semibold text-red-900 mb-2">How to fix:</p>
          <ol className="list-decimal list-inside space-y-1">
            {howToFix.map((step, index) => (
              <li key={index} className="text-sm text-red-800">
                {step}
              </li>
            ))}
          </ol>
        </div>

        {/* Actions */}
        {(retryable || additionalActions.length > 0 || enableBugReport) && (
          <div className="flex flex-wrap gap-2 pt-2">
            {retryable && onRetry && (
              <button
                onClick={onRetry}
                className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white font-semibold rounded transition-colors"
              >
                Retry
              </button>
            )}
            {additionalActions.map((action, index) => (
              <button
                key={index}
                onClick={action.onClick}
                className={`px-4 py-2 font-semibold rounded transition-colors ${
                  action.variant === 'primary'
                    ? 'bg-red-600 hover:bg-red-700 text-white'
                    : 'bg-white hover:bg-red-50 text-red-700 border border-red-300'
                }`}
              >
                {action.label}
              </button>
            ))}
            {enableBugReport && (
              <button
                onClick={() => setShowBugReport(true)}
                className="px-4 py-2 bg-white hover:bg-red-50 text-red-700 border border-red-300 font-semibold rounded transition-colors flex items-center gap-2"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Report Bug
              </button>
            )}
          </div>
        )}
      </div>
    </div>

      <BugReportDialog
        isOpen={showBugReport}
        onClose={() => setShowBugReport(false)}
        prefilledDescription={bugReportDescription}
      />
    </>
  )
}
