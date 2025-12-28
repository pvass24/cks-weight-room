'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import AppLayout from '@/components/AppLayout'
import { getExercises } from '@/lib/api'
import type { Exercise } from '@/types/exercise'
import { CategoryLabels } from '@/types/exercise'
import {
  FlaskConical,
  Clock,
  CheckCircle,
  ArrowRight
} from 'lucide-react'

type DifficultyFilter = 'all' | 'beginner' | 'intermediate' | 'advanced'

export default function ExercisesPage() {
  const router = useRouter()
  const [exercises, setExercises] = useState<Exercise[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [difficultyFilter, setDifficultyFilter] = useState<DifficultyFilter>('all')
  const [provisioningSlug, setProvisioningSlug] = useState<string | null>(null)

  useEffect(() => {
    let mounted = true
    const timeoutId = setTimeout(() => {
      if (mounted && loading) {
        console.error('Exercise loading timed out')
        setError('Loading timed out. Please refresh the page.')
        setLoading(false)
      }
    }, 10000)

    async function loadExercises() {
      try {
        const activationResponse = await fetch('/api/activation/status')
        const activationStatus = await activationResponse.json()

        if (!mounted) return

        if (!activationStatus.isActivated) {
          router.push('/activate')
          return
        }

        const data = await getExercises()
        if (mounted) {
          setExercises(data)
          clearTimeout(timeoutId)
        }
      } catch (err) {
        console.error('Failed to load exercises:', err)
        if (mounted) {
          setError(err instanceof Error ? err.message : 'Failed to load exercises')
        }
      } finally {
        if (mounted) {
          setLoading(false)
        }
      }
    }

    loadExercises()

    return () => {
      mounted = false
      clearTimeout(timeoutId)
    }
  }, [router])

  const handleStartLab = async (exerciseSlug: string) => {
    setProvisioningSlug(exerciseSlug)

    try {
      const response = await fetch('/api/cluster/provision', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ exerciseSlug }),
      })

      const data = await response.json()

      if (data.success) {
        // Redirect to practice view
        router.push(`/practice/${exerciseSlug}`)
      } else {
        alert(`Failed to provision cluster: ${data.error}`)
        setProvisioningSlug(null)
      }
    } catch (err) {
      alert(`Error: ${err instanceof Error ? err.message : 'Failed to provision cluster'}`)
      setProvisioningSlug(null)
    }
  }

  const filteredExercises = difficultyFilter === 'all'
    ? exercises
    : exercises.filter(ex => ex.difficulty.toLowerCase() === difficultyFilter)

  const difficultyBadgeStyles: Record<string, string> = {
    beginner: 'bg-green-50 text-green-700 border-green-200',
    intermediate: 'bg-yellow-50 text-yellow-700 border-yellow-200',
    advanced: 'bg-red-50 text-red-700 border-red-200'
  }

  if (loading) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center min-h-[60vh]">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
            <p className="text-gray-600">Loading practice labs...</p>
          </div>
        </div>
      </AppLayout>
    )
  }

  if (error) {
    return (
      <AppLayout>
        <div className="flex items-center justify-center min-h-[60vh]">
          <div className="text-center">
            <div className="text-red-600 text-xl font-semibold mb-2">Failed to load practice labs</div>
            <p className="text-gray-600">{error}</p>
          </div>
        </div>
      </AppLayout>
    )
  }

  return (
    <AppLayout>
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Practice Labs</h1>
          <p className="text-gray-600">
            Hands-on exercises to reinforce your Kubernetes security skills
          </p>
        </div>

        {/* Difficulty Filter */}
        <div className="flex gap-2 mb-6">
          {(['all', 'beginner', 'intermediate', 'advanced'] as DifficultyFilter[]).map((level) => (
            <button
              key={level}
              onClick={() => setDifficultyFilter(level)}
              className={`px-4 py-2 rounded-lg font-medium text-sm transition-colors ${
                difficultyFilter === level
                  ? 'bg-blue-600 text-white'
                  : 'bg-white text-gray-700 border border-gray-200 hover:bg-gray-50'
              }`}
            >
              {level === 'all' ? 'All Labs' : level.charAt(0).toUpperCase() + level.slice(1)}
            </button>
          ))}
        </div>

        {/* Labs Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {filteredExercises.map((exercise) => (
            <div
              key={exercise.slug}
              className="bg-white rounded-xl border border-gray-200 p-6 hover:shadow-lg transition-shadow group"
            >
              <div className="flex items-start gap-4">
                {/* Icon */}
                <div className="bg-blue-50 p-3 rounded-lg">
                  <FlaskConical className="w-6 h-6 text-blue-600" />
                </div>

                {/* Content */}
                <div className="flex-1">
                  <h3 className="text-lg font-semibold text-gray-900 mb-2 group-hover:text-blue-600 transition-colors">
                    {exercise.title}
                  </h3>

                  <p className="text-sm text-gray-600 mb-4 line-clamp-2">
                    {exercise.description}
                  </p>

                  {/* Metadata */}
                  <div className="flex flex-wrap items-center gap-3 mb-4">
                    <span
                      className={`px-3 py-1 rounded-full text-xs font-semibold border ${
                        difficultyBadgeStyles[exercise.difficulty.toLowerCase()] ||
                        'bg-gray-50 text-gray-700 border-gray-200'
                      }`}
                    >
                      {exercise.difficulty}
                    </span>

                    <div className="flex items-center gap-1.5 text-sm text-gray-600">
                      <Clock className="w-4 h-4" />
                      <span>{exercise.estimatedMinutes} min</span>
                    </div>

                    <div className="text-sm text-gray-600">
                      {CategoryLabels[exercise.category] || exercise.category}
                    </div>
                  </div>

                  {/* Action Button */}
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      handleStartLab(exercise.slug)
                    }}
                    disabled={provisioningSlug === exercise.slug}
                    className="w-full bg-blue-600 text-white font-semibold py-2.5 px-4 rounded-lg hover:bg-blue-700 disabled:bg-gray-400 transition-colors flex items-center justify-center gap-2"
                  >
                    {provisioningSlug === exercise.slug ? (
                      <>
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                        <span>Starting...</span>
                      </>
                    ) : (
                      <>
                        <span>Start Lab</span>
                        <ArrowRight className="w-4 h-4" />
                      </>
                    )}
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>

        {filteredExercises.length === 0 && (
          <div className="text-center py-12">
            <p className="text-gray-500">No practice labs found for this difficulty level.</p>
          </div>
        )}
      </div>
    </AppLayout>
  )
}
