import type { ValidationResponse, APIError, InitializeResponse, DatabaseStatus } from '@/types/setup'
import type { ExercisesResponse, Exercise } from '@/types/exercise'

/**
 * Validates prerequisites by calling the backend API
 * @returns Promise with validation response
 * @throws APIError if request fails
 */
export async function validatePrerequisites(): Promise<ValidationResponse> {
  try {
    const response = await fetch('/api/setup/validate')

    if (!response.ok) {
      throw {
        code: 'HTTP_ERROR',
        message: `HTTP ${response.status}: ${response.statusText}`,
      } as APIError
    }

    const data: ValidationResponse = await response.json()
    return data
  } catch (error) {
    if ((error as APIError).code) {
      throw error
    }

    throw {
      code: 'NETWORK_ERROR',
      message: error instanceof Error ? error.message : 'Network request failed',
    } as APIError
  }
}

/**
 * Initializes the database
 * @returns Promise with initialization response
 * @throws APIError if request fails
 */
export async function initializeDatabase(): Promise<InitializeResponse> {
  try {
    const response = await fetch('/api/setup/initialize', {
      method: 'POST',
    })

    const data: InitializeResponse = await response.json()

    if (!response.ok) {
      throw {
        code: data.errorCode || 'HTTP_ERROR',
        message: data.message || `HTTP ${response.status}: ${response.statusText}`,
      } as APIError
    }

    return data
  } catch (error) {
    if ((error as APIError).code) {
      throw error
    }

    throw {
      code: 'NETWORK_ERROR',
      message: error instanceof Error ? error.message : 'Network request failed',
    } as APIError
  }
}

/**
 * Gets database initialization status
 * @returns Promise with database status
 * @throws APIError if request fails
 */
export async function getDatabaseStatus(): Promise<DatabaseStatus> {
  try {
    const response = await fetch('/api/setup/db-status')

    if (!response.ok) {
      throw {
        code: 'HTTP_ERROR',
        message: `HTTP ${response.status}: ${response.statusText}`,
      } as APIError
    }

    const data: DatabaseStatus = await response.json()
    return data
  } catch (error) {
    if ((error as APIError).code) {
      throw error
    }

    throw {
      code: 'NETWORK_ERROR',
      message: error instanceof Error ? error.message : 'Network request failed',
    } as APIError
  }
}

/**
 * Gets all exercises, optionally filtered by category
 * @param category Optional category filter
 * @returns Promise with exercises array
 * @throws APIError if request fails
 */
export async function getExercises(category?: string): Promise<Exercise[]> {
  try {
    const url = category ? `/api/exercises?category=${encodeURIComponent(category)}` : '/api/exercises'
    const response = await fetch(url)

    if (!response.ok) {
      throw {
        code: 'HTTP_ERROR',
        message: `HTTP ${response.status}: ${response.statusText}`,
      } as APIError
    }

    const data: ExercisesResponse = await response.json()

    if (!data.success) {
      throw {
        code: data.errorCode || 'API_ERROR',
        message: data.message || 'Failed to fetch exercises',
      } as APIError
    }

    return data.exercises || []
  } catch (error) {
    if ((error as APIError).code) {
      throw error
    }

    throw {
      code: 'NETWORK_ERROR',
      message: error instanceof Error ? error.message : 'Network request failed',
    } as APIError
  }
}

/**
 * Gets a single exercise by slug
 * @param slug Exercise slug identifier
 * @returns Promise with exercise
 * @throws APIError if request fails
 */
export async function getExerciseBySlug(slug: string): Promise<Exercise> {
  try {
    const response = await fetch(`/api/exercises/${encodeURIComponent(slug)}`)

    if (!response.ok) {
      throw {
        code: 'HTTP_ERROR',
        message: `HTTP ${response.status}: ${response.statusText}`,
      } as APIError
    }

    const data: ExercisesResponse = await response.json()

    if (!data.success || !data.exercise) {
      throw {
        code: data.errorCode || 'NOT_FOUND',
        message: data.message || 'Exercise not found',
      } as APIError
    }

    return data.exercise
  } catch (error) {
    if ((error as APIError).code) {
      throw error
    }

    throw {
      code: 'NETWORK_ERROR',
      message: error instanceof Error ? error.message : 'Network request failed',
    } as APIError
  }
}
