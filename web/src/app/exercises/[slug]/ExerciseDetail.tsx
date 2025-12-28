'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import type { Exercise } from '@/types/exercise'
import { CategoryLabels, DifficultyColors } from '@/types/exercise'
import ActionableError from '@/components/ActionableError'

interface ActionableErrorData {
  code: string
  what: string
  why: string
  howToFix: string[]
  retryable: boolean
  context?: Record<string, string>
}

export default function ExerciseDetail({ exercise }: { exercise: Exercise }) {
  const router = useRouter()
  const [showHints, setShowHints] = useState(false)
  const [showSolution, setShowSolution] = useState(false)
  const [revealedHints, setRevealedHints] = useState<number>(0)
  const [provisioning, setProvisioning] = useState(false)
  const [provisionError, setProvisionError] = useState<ActionableErrorData | null>(null)

  const handleStartScenario = async () => {
    setProvisioning(true)
    setProvisionError(null)

    try {
      const response = await fetch('/api/cluster/provision', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ exerciseSlug: exercise.slug }),
      })

      const data = await response.json()

      if (!data.success) {
        // Use actionableError if available, otherwise create basic error
        if (data.actionableError) {
          setProvisionError(data.actionableError)
        } else {
          setProvisionError({
            code: 'UNKNOWN_ERROR',
            what: 'Cluster provisioning failed',
            why: data.error || 'An unknown error occurred',
            howToFix: ['Try the operation again', 'Contact support if the issue persists'],
            retryable: true,
          })
        }
        setProvisioning(false)
        return
      }

      // Redirect to practice view
      router.push(`/practice/${exercise.slug}`)
    } catch (err) {
      setProvisionError({
        code: 'NETWORK_ERROR',
        what: 'Network request failed',
        why: err instanceof Error ? err.message : 'Failed to connect to server',
        howToFix: ['Check your connection', 'Refresh the page and try again'],
        retryable: true,
      })
      setProvisioning(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto px-4 py-8">
        {/* Back Button */}
        <button
          onClick={() => router.push('/exercises')}
          className="mb-6 flex items-center text-gray-600 hover:text-gray-900"
        >
          <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          Back to Exercises
        </button>

        {/* Exercise Header */}
        <div className="bg-white rounded-lg shadow-md p-8 mb-6">
          <div className="flex items-start justify-between mb-4">
            <h1 className="text-3xl font-bold text-gray-900">{exercise.title}</h1>
            <div className="flex items-center gap-2 ml-4 flex-shrink-0">
              <span className={`px-3 py-1 rounded-full text-sm font-medium ${DifficultyColors[exercise.difficulty]}`}>
                {exercise.difficulty}
              </span>
              <span className="px-3 py-1 rounded-full text-sm font-medium bg-blue-50 text-blue-600">
                {exercise.points} pts
              </span>
            </div>
          </div>

          <div className="flex items-center gap-4 text-sm text-gray-500 mb-6 flex-wrap">
            <span className="flex items-center gap-1">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
              </svg>
              {CategoryLabels[exercise.category] || exercise.category}
            </span>
            <span className="flex items-center gap-1">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              ~{exercise.estimatedMinutes} min
            </span>
          </div>

          <div className="prose max-w-none">
            <h2 className="text-xl font-semibold mb-3">Description</h2>
            <p className="text-gray-700 leading-relaxed whitespace-pre-wrap">{exercise.description}</p>
          </div>

          {exercise.prerequisites.length > 0 && (
            <div className="mt-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
              <h3 className="font-semibold text-yellow-800 mb-2">Prerequisites</h3>
              <ul className="list-disc list-inside text-yellow-700">
                {exercise.prerequisites.map((prereq, idx) => (
                  <li key={idx}>{prereq}</li>
                ))}
              </ul>
            </div>
          )}
        </div>

        {/* Hints Section */}
        {exercise.hints.length > 0 && (
          <div className="bg-white rounded-lg shadow-md p-8 mb-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-2xl font-semibold text-gray-900">Hints</h2>
              <button
                onClick={() => {
                  setShowHints(!showHints)
                  if (!showHints) setRevealedHints(0)
                }}
                className="text-blue-600 hover:text-blue-700 font-medium"
              >
                {showHints ? 'Hide' : 'Show'} Hints
              </button>
            </div>

            {showHints && (
              <div className="space-y-3">
                {revealedHints === 0 ? (
                  <div className="text-center py-4">
                    <p className="text-gray-600 mb-4">
                      There {exercise.hints.length === 1 ? 'is' : 'are'} {exercise.hints.length} hint
                      {exercise.hints.length === 1 ? '' : 's'} available.
                    </p>
                    <div className="flex gap-3 justify-center">
                      <button
                        onClick={() => setRevealedHints(1)}
                        className="bg-blue-600 hover:bg-blue-700 text-white font-semibold py-2 px-4 rounded-lg"
                      >
                        Reveal one hint at a time
                      </button>
                      <button
                        onClick={() => setRevealedHints(exercise.hints.length)}
                        className="bg-gray-600 hover:bg-gray-700 text-white font-semibold py-2 px-4 rounded-lg"
                      >
                        Show all hints
                      </button>
                    </div>
                  </div>
                ) : (
                  <>
                    {exercise.hints.slice(0, revealedHints).map((hint, idx) => (
                      <div key={idx} className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
                        <div className="flex items-start">
                          <span className="flex-shrink-0 w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm font-semibold mr-3">
                            {idx + 1}
                          </span>
                          <p className="text-gray-700">{hint}</p>
                        </div>
                      </div>
                    ))}
                    {revealedHints < exercise.hints.length && (
                      <button
                        onClick={() => setRevealedHints(revealedHints + 1)}
                        className="text-blue-600 hover:text-blue-700 font-medium"
                      >
                        Show next hint ({revealedHints} of {exercise.hints.length} revealed)
                      </button>
                    )}
                  </>
                )}
              </div>
            )}
          </div>
        )}

        {/* Solution Section */}
        <div className="bg-white rounded-lg shadow-md p-8">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-2xl font-semibold text-gray-900">Solution</h2>
            <button
              onClick={() => setShowSolution(!showSolution)}
              className="bg-green-600 hover:bg-green-700 text-white font-semibold py-2 px-4 rounded-lg"
            >
              {showSolution ? 'Hide' : 'Show'} Solution
            </button>
          </div>

          {showSolution && (
            <div className="p-6 bg-green-50 border border-green-200 rounded-lg">
              <p className="text-gray-800 whitespace-pre-wrap font-mono text-sm leading-relaxed">
                {exercise.solution}
              </p>
            </div>
          )}

          {!showSolution && (
            <p className="text-gray-500 italic">
              Try to solve the exercise yourself before viewing the solution.
            </p>
          )}
        </div>

        {/* Start Scenario Button */}
        <div className="bg-white rounded-lg shadow-md p-8 mb-6">
          <h2 className="text-2xl font-semibold text-gray-900 mb-4">Ready to Practice?</h2>
          <p className="text-gray-600 mb-6">
            Click below to provision a KIND cluster and start practicing this exercise in a real Kubernetes environment.
          </p>

          {provisionError && (
            <div className="mb-4">
              <ActionableError
                code={provisionError.code}
                what={provisionError.what}
                why={provisionError.why}
                howToFix={provisionError.howToFix}
                retryable={provisionError.retryable}
                context={provisionError.context}
                onRetry={handleStartScenario}
                onDismiss={() => setProvisionError(null)}
              />
            </div>
          )}

          <button
            onClick={handleStartScenario}
            disabled={provisioning}
            className="w-full bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white font-bold py-4 px-8 rounded-lg text-lg transition-colors flex items-center justify-center gap-3"
          >
            {provisioning ? (
              <>
                <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
                Provisioning Cluster...
              </>
            ) : (
              <>
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Start Scenario
              </>
            )}
          </button>
        </div>

        {/* Action Buttons */}
        <div className="mt-8 flex gap-4">
          <button
            onClick={() => router.push('/exercises')}
            className="flex-1 bg-gray-600 hover:bg-gray-700 text-white font-semibold py-3 px-6 rounded-lg"
          >
            Back to All Exercises
          </button>
        </div>
      </div>
    </div>
  )
}
