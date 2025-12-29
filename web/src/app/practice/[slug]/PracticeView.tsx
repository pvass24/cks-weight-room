'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { getExerciseBySlug } from '@/lib/api'
import type { Exercise } from '@/types/exercise'
import { DifficultyColors } from '@/types/exercise'
import RightPanel from '@/components/RightPanel'
import Timer from '@/components/Timer'
import ActionableError from '@/components/ActionableError'

interface ActionableErrorData {
  code: string
  what: string
  why: string
  howToFix: string[]
  retryable: boolean
  context?: Record<string, string>
}

interface ClusterStatus {
  name: string
  exerciseSlug: string
  status: 'provisioning' | 'ready' | 'error' | 'not_found'
  createdAt?: string
  errorMessage?: string
  kubeconfigContext?: string
}

export default function PracticeView({ slug }: { slug: string }) {
  const router = useRouter()
  const [exercise, setExercise] = useState<Exercise | null>(null)
  const [clusterStatus, setClusterStatus] = useState<ClusterStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showHints, setShowHints] = useState(false)
  const [revealedHints, setRevealedHints] = useState<number>(0)
  const [personalBest, setPersonalBest] = useState<number | undefined>(undefined)
  const [resetting, setResetting] = useState(false)
  const [timerKey, setTimerKey] = useState(0)
  const [validating, setValidating] = useState(false)
  const [selectedNode, setSelectedNode] = useState<string>('')
  const [validationResult, setValidationResult] = useState<{
    passed: boolean
    score: number
    feedback: string
    details?: string[]
  } | null>(null)
  const [resetError, setResetError] = useState<ActionableErrorData | null>(null)

  useEffect(() => {
    const fetchData = async () => {
      try {
        // Fetch exercise details
        const exerciseData = await getExerciseBySlug(slug)
        setExercise(exerciseData)

        // Check cluster status
        const clusterRes = await fetch(`/api/cluster/status/${slug}`)
        if (clusterRes.ok) {
          const clusterData = await clusterRes.json()
          setClusterStatus(clusterData.cluster)
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load practice environment')
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [slug])

  // Poll cluster status when provisioning
  useEffect(() => {
    if (!clusterStatus || clusterStatus.status !== 'provisioning') return

    const pollInterval = setInterval(async () => {
      try {
        const clusterRes = await fetch(`/api/cluster/status/${slug}`)
        if (clusterRes.ok) {
          const clusterData = await clusterRes.json()
          setClusterStatus(clusterData.cluster)

          // Stop polling when cluster is ready or errored
          if (clusterData.cluster.status === 'ready' || clusterData.cluster.status === 'error') {
            clearInterval(pollInterval)
          }
        }
      } catch (err) {
        console.error('Failed to poll cluster status:', err)
      }
    }, 3000) // Poll every 3 seconds

    return () => clearInterval(pollInterval)
  }, [slug, clusterStatus?.status])

  const handleEndSession = async () => {
    if (!confirm('Are you sure you want to end this practice session? The cluster will be deleted.')) {
      return
    }

    try {
      const response = await fetch(`/api/cluster/${slug}`, {
        method: 'DELETE',
      })

      if (response.ok) {
        router.push(`/exercises/${slug}`)
      } else {
        const data = await response.json()
        alert(`Failed to delete cluster: ${data.error}`)
      }
    } catch (err) {
      alert(`Error: ${err instanceof Error ? err.message : 'Failed to end session'}`)
    }
  }

  const handleReset = async () => {
    if (!confirm('Reset this scenario? This will delete and recreate the cluster, clearing all your work.')) {
      return
    }

    setResetting(true)
    setResetError(null)

    try {
      // Delete existing cluster
      const deleteResponse = await fetch(`/api/cluster/${slug}`, {
        method: 'DELETE',
      })

      if (!deleteResponse.ok) {
        const data = await deleteResponse.json()
        throw new Error(data.error || 'Failed to delete cluster')
      }

      // Wait a moment for cleanup
      await new Promise(resolve => setTimeout(resolve, 2000))

      // Provision new cluster
      const provisionResponse = await fetch('/api/cluster/provision', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ exerciseSlug: slug }),
      })

      const data = await provisionResponse.json()

      if (!data.success) {
        // Use actionableError if available
        if (data.actionableError) {
          setResetError(data.actionableError)
        } else {
          setResetError({
            code: 'RESET_FAILED',
            what: 'Cluster reset failed',
            why: data.error || 'Failed to provision new cluster',
            howToFix: ['Try resetting again', 'Contact support if the issue persists'],
            retryable: true,
          })
        }
        setResetting(false)
        return
      }

      // Reset timer by changing key (forces remount)
      setTimerKey(prev => prev + 1)

      // Update cluster status
      setClusterStatus(data.cluster)

      // Reload the page to reconnect terminal
      window.location.reload()
    } catch (err) {
      setResetError({
        code: 'RESET_ERROR',
        what: 'Reset operation failed',
        why: err instanceof Error ? err.message : 'Unknown error',
        howToFix: ['Try resetting again', 'Refresh the page and try again'],
        retryable: true,
      })
      setResetting(false)
    }
  }

  const handleValidate = async () => {
    setValidating(true)
    setValidationResult(null)

    try {
      const response = await fetch(`/api/validate/${slug}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          clusterName: clusterStatus?.name,
        }),
      })

      const data = await response.json()

      if (!response.ok) {
        throw new Error(data.error || 'Validation failed')
      }

      setValidationResult(data)
    } catch (err) {
      alert(`Validation error: ${err instanceof Error ? err.message : 'Unknown error'}`)
    } finally {
      setValidating(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading practice environment...</p>
        </div>
      </div>
    )
  }

  if (error || !exercise) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="bg-white rounded-lg shadow-md p-8 max-w-md">
          <h2 className="text-xl font-bold text-red-600 mb-4">Error</h2>
          <p className="text-gray-700 mb-6">{error || 'Exercise not found'}</p>
          <button
            onClick={() => router.push('/exercises')}
            className="w-full bg-gray-600 hover:bg-gray-700 text-white font-semibold py-2 px-4 rounded-lg"
          >
            Back to Exercises
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header Bar */}
      <div className="bg-white border-b border-gray-200 px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h1 className="text-2xl font-bold text-gray-900">{exercise.title}</h1>
            <span className={`px-3 py-1 rounded-full text-sm font-medium ${DifficultyColors[exercise.difficulty]}`}>
              {exercise.difficulty}
            </span>
            <span className="px-3 py-1 rounded-full text-sm font-medium bg-blue-50 text-blue-600">
              {exercise.points} pts
            </span>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={handleReset}
              disabled={resetting}
              className="bg-orange-600 hover:bg-orange-700 disabled:bg-gray-400 text-white font-semibold py-2 px-4 rounded-lg flex items-center gap-2"
            >
              {resetting ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                  Resetting...
                </>
              ) : (
                <>
                  <svg className="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Reset
                </>
              )}
            </button>
            <button
              onClick={handleEndSession}
              className="bg-red-600 hover:bg-red-700 text-white font-semibold py-2 px-4 rounded-lg"
            >
              End Session
            </button>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex h-[calc(100vh-73px)]">
        {/* Left Panel - Exercise Details & Hints */}
        <div className="w-1/3 bg-white border-r border-gray-200 overflow-y-auto">
          <div className="p-6">
            {/* Timer */}
            <div className="mb-6">
              <Timer key={timerKey} exerciseSlug={slug} personalBest={personalBest} />
            </div>

            {/* Reset Error */}
            {resetError && (
              <div className="mb-6">
                <ActionableError
                  code={resetError.code}
                  what={resetError.what}
                  why={resetError.why}
                  howToFix={resetError.howToFix}
                  retryable={resetError.retryable}
                  context={resetError.context}
                  onRetry={handleReset}
                  onDismiss={() => setResetError(null)}
                />
              </div>
            )}

            {/* Cluster Status */}
            {clusterStatus && (
              <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
                <h3 className="font-semibold text-blue-900 mb-2">Cluster Status</h3>
                <div className="flex items-center gap-2">
                  <span className={`inline-block w-3 h-3 rounded-full ${
                    clusterStatus.status === 'ready' ? 'bg-green-500' :
                    clusterStatus.status === 'provisioning' ? 'bg-yellow-500' :
                    clusterStatus.status === 'error' ? 'bg-red-500' :
                    'bg-gray-500'
                  }`}></span>
                  <span className="text-sm font-medium text-blue-800">
                    {clusterStatus.status === 'ready' ? 'Ready' :
                     clusterStatus.status === 'provisioning' ? 'Provisioning...' :
                     clusterStatus.status === 'error' ? 'Error' :
                     'Not Found'}
                  </span>
                </div>
                {clusterStatus.kubeconfigContext && (
                  <p className="text-xs text-blue-700 mt-2">
                    Context: <code className="bg-white px-1 py-0.5 rounded">{clusterStatus.kubeconfigContext}</code>
                  </p>
                )}
                {clusterStatus.errorMessage && (
                  <p className="text-xs text-red-700 mt-2">{clusterStatus.errorMessage}</p>
                )}
              </div>
            )}

            {/* Exercise Description */}
            <div className="mb-6">
              <h2 className="text-xl font-semibold mb-3 text-gray-900">Description</h2>
              <p className="text-gray-700 leading-relaxed whitespace-pre-wrap">{exercise.description}</p>
            </div>

            {/* Prerequisites */}
            {exercise.prerequisites.length > 0 && (
              <div className="mb-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
                <h3 className="font-semibold text-yellow-800 mb-2">Prerequisites</h3>
                <ul className="list-disc list-inside text-yellow-700 text-sm">
                  {exercise.prerequisites.map((prereq, idx) => (
                    <li key={idx}>{prereq}</li>
                  ))}
                </ul>
              </div>
            )}

            {/* Hints Section */}
            {exercise.hints.length > 0 && (
              <div className="mb-6">
                <div className="flex items-center justify-between mb-3">
                  <h2 className="text-xl font-semibold text-gray-900">Hints</h2>
                  <button
                    onClick={() => {
                      setShowHints(!showHints)
                      if (!showHints) setRevealedHints(0)
                    }}
                    className="text-blue-600 hover:text-blue-700 font-medium text-sm"
                  >
                    {showHints ? 'Hide' : 'Show'} Hints
                  </button>
                </div>

                {showHints && (
                  <div className="space-y-3">
                    {revealedHints === 0 ? (
                      <div className="text-center py-4 bg-gray-50 rounded-lg">
                        <p className="text-gray-600 mb-4 text-sm">
                          There {exercise.hints.length === 1 ? 'is' : 'are'} {exercise.hints.length} hint
                          {exercise.hints.length === 1 ? '' : 's'} available.
                        </p>
                        <div className="flex gap-3 justify-center">
                          <button
                            onClick={() => setRevealedHints(1)}
                            className="bg-blue-600 hover:bg-blue-700 text-white font-semibold py-2 px-4 rounded-lg text-sm"
                          >
                            Reveal one hint
                          </button>
                          <button
                            onClick={() => setRevealedHints(exercise.hints.length)}
                            className="bg-gray-600 hover:bg-gray-700 text-white font-semibold py-2 px-4 rounded-lg text-sm"
                          >
                            Show all
                          </button>
                        </div>
                      </div>
                    ) : (
                      <>
                        {exercise.hints.slice(0, revealedHints).map((hint, idx) => (
                          <div key={idx} className="p-3 bg-blue-50 border border-blue-200 rounded-lg">
                            <div className="flex items-start">
                              <span className="flex-shrink-0 w-6 h-6 bg-blue-600 text-white rounded-full flex items-center justify-center text-xs font-semibold mr-3">
                                {idx + 1}
                              </span>
                              <p className="text-gray-700 text-sm">{hint}</p>
                            </div>
                          </div>
                        ))}
                        {revealedHints < exercise.hints.length && (
                          <button
                            onClick={() => setRevealedHints(revealedHints + 1)}
                            className="text-blue-600 hover:text-blue-700 font-medium text-sm"
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

            {/* Validation Section */}
            <div className="mb-6">
              <button
                onClick={handleValidate}
                disabled={validating || clusterStatus?.status !== 'ready'}
                className="w-full bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white font-bold py-3 px-6 rounded-lg flex items-center justify-center gap-2"
              >
                {validating ? (
                  <>
                    <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
                    Validating Solution...
                  </>
                ) : (
                  <>
                    <svg className="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    Submit Solution
                  </>
                )}
              </button>

              {validationResult && (
                <div className={`mt-4 p-4 rounded-lg border ${
                  validationResult.passed
                    ? 'bg-green-50 border-green-200'
                    : 'bg-red-50 border-red-200'
                }`}>
                  <div className="flex items-start gap-3">
                    {validationResult.passed ? (
                      <svg className="w-6 h-6 text-green-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    ) : (
                      <svg className="w-6 h-6 text-red-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    )}
                    <div className="flex-1">
                      <h3 className={`font-semibold ${
                        validationResult.passed ? 'text-green-900' : 'text-red-900'
                      }`}>
                        {validationResult.passed ? 'Solution Correct!' : 'Solution Incorrect'}
                      </h3>
                      <p className={`text-sm mt-1 ${
                        validationResult.passed ? 'text-green-700' : 'text-red-700'
                      }`}>
                        {validationResult.feedback}
                      </p>
                      <div className="mt-2">
                        <span className={`text-lg font-bold ${
                          validationResult.passed ? 'text-green-600' : 'text-red-600'
                        }`}>
                          Score: {validationResult.score}/{exercise.points}
                        </span>
                      </div>
                      {validationResult.details && validationResult.details.length > 0 && (
                        <ul className="mt-3 space-y-1">
                          {validationResult.details.map((detail, idx) => (
                            <li key={idx} className={`text-sm ${
                              validationResult.passed ? 'text-green-700' : 'text-red-700'
                            }`}>
                              • {detail}
                            </li>
                          ))}
                        </ul>
                      )}
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Right Panel - Terminal & IDE */}
        <RightPanel
          exerciseSlug={slug}
          clusterReady={clusterStatus?.status === 'ready'}
          selectedNode={selectedNode}
          onNodeChange={setSelectedNode}
        />
      </div>
    </div>
  )
}
