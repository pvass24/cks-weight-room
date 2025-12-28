'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { getExerciseBySlug } from '@/lib/api'
import type { Exercise } from '@/types/exercise'
import ExerciseDetail from './ExerciseDetail'

export default function ExerciseClientPage({ slug }: { slug: string }) {
  const router = useRouter()
  const [exercise, setExercise] = useState<Exercise | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let mounted = true
    const timeoutId = setTimeout(() => {
      if (mounted && loading) {
        console.error('Exercise loading timed out')
        setError('Loading timed out. Please try again.')
        setLoading(false)
      }
    }, 10000)

    async function loadExercise() {
      try {
        console.log(`Loading exercise: ${slug}`)
        setLoading(true)
        setError(null)
        const data = await getExerciseBySlug(slug)
        console.log('Exercise loaded:', data.title)
        if (mounted) {
          setExercise(data)
          clearTimeout(timeoutId)
        }
      } catch (err) {
        console.error('Failed to load exercise:', err)
        if (mounted) {
          setError(err instanceof Error ? err.message : 'Failed to load exercise')
        }
      } finally {
        if (mounted) {
          setLoading(false)
        }
      }
    }

    loadExercise()

    return () => {
      mounted = false
      clearTimeout(timeoutId)
    }
  }, [slug])

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading exercise...</p>
        </div>
      </div>
    )
  }

  if (error || !exercise) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center max-w-md">
          <div className="text-red-600 text-xl mb-4">Failed to load exercise</div>
          <p className="text-gray-600 mb-4">{error || 'Exercise not found'}</p>
          <button
            onClick={() => router.push('/exercises')}
            className="bg-blue-600 hover:bg-blue-700 text-white font-semibold py-2 px-6 rounded-lg"
          >
            Back to Exercises
          </button>
        </div>
      </div>
    )
  }

  return <ExerciseDetail exercise={exercise} />
}
