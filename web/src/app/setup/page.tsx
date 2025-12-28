'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { validatePrerequisites, initializeDatabase } from '@/lib/api'
import type { ValidationResponse, CheckResult } from '@/types/setup'

export default function SetupPage() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [validation, setValidation] = useState<ValidationResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [initializing, setInitializing] = useState(false)

  useEffect(() => {
    async function checkPrerequisites() {
      try {
        const result = await validatePrerequisites()
        setValidation(result)

        // If prerequisites pass, automatically initialize database
        if (result.success) {
          setInitializing(true)
          try {
            await initializeDatabase()
            // Wait a moment then redirect to home
            setTimeout(() => {
              router.push('/')
            }, 2000)
          } catch (initErr) {
            setError(initErr instanceof Error ? initErr.message : 'Failed to initialize database')
            setInitializing(false)
          }
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to validate prerequisites')
      } finally {
        setLoading(false)
      }
    }

    checkPrerequisites()
  }, [router])

  if (loading) {
    return (
      <main className="flex min-h-screen flex-col items-center justify-center p-24">
        <div className="text-center">
          <h1 className="text-4xl font-bold mb-8">CKS Weight Room</h1>
          <p className="text-xl text-gray-600">Checking prerequisites...</p>
        </div>
      </main>
    )
  }

  if (error) {
    return (
      <main className="flex min-h-screen flex-col items-center justify-center p-24">
        <div className="text-center">
          <h1 className="text-4xl font-bold mb-8 text-red-600">Error</h1>
          <p className="text-xl text-gray-600">{error}</p>
        </div>
      </main>
    )
  }

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24">
      <div className="max-w-2xl w-full">
        <h1 className="text-4xl font-bold mb-8 text-center">CKS Weight Room Setup</h1>

        <div className="bg-white rounded-lg shadow-lg p-8">
          <h2 className="text-2xl font-semibold mb-6">Prerequisite Validation</h2>

          <div className="space-y-4">
            {validation?.checks.map((check: CheckResult) => (
              <div
                key={check.name}
                className="flex items-start p-4 border rounded-lg"
              >
                <div className="flex-shrink-0 mr-4">
                  {check.passed ? (
                    <svg
                      className="w-6 h-6 text-green-500"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M5 13l4 4L19 7"
                      />
                    </svg>
                  ) : (
                    <svg
                      className="w-6 h-6 text-red-500"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M6 18L18 6M6 6l12 12"
                      />
                    </svg>
                  )}
                </div>
                <div className="flex-1">
                  <h3 className="font-semibold text-lg">{check.name}</h3>
                  {check.message && (
                    <p className="text-gray-600 mt-2">{check.message}</p>
                  )}
                </div>
              </div>
            ))}
          </div>

          {validation?.success ? (
            <div className="mt-8 p-4 bg-green-50 border border-green-200 rounded-lg">
              <p className="text-green-800 font-semibold text-center">
                ✓ All prerequisites validated
              </p>
              {initializing ? (
                <p className="text-green-600 text-center mt-2">
                  Initializing database...
                </p>
              ) : (
                <p className="text-green-600 text-center mt-2">
                  Redirecting to application...
                </p>
              )}
            </div>
          ) : (
            <div className="mt-8 p-4 bg-red-50 border border-red-200 rounded-lg">
              <p className="text-red-800 font-semibold">
                Prerequisites not met
              </p>
              {validation?.message && (
                <p className="text-red-600 mt-2">{validation.message}</p>
              )}
              <div className="mt-4">
                <p className="text-sm text-gray-600 font-semibold mb-2">
                  Installation Links:
                </p>
                <ul className="text-sm text-gray-600 space-y-1">
                  <li>
                    <a
                      href="https://www.docker.com/products/docker-desktop"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 hover:underline"
                    >
                      Docker Desktop
                    </a>
                  </li>
                  <li>
                    <a
                      href="https://kind.sigs.k8s.io/docs/user/quick-start/"
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 hover:underline"
                    >
                      KIND Installation Guide
                    </a>
                  </li>
                </ul>
              </div>
            </div>
          )}
        </div>
      </div>
    </main>
  )
}
