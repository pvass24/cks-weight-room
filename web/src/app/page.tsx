'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

export default function Home() {
  const router = useRouter()

  useEffect(() => {
    async function checkSetup() {
      try {
        // Check if database is initialized
        const dbResponse = await fetch('/api/setup/db-status')
        const dbStatus = await dbResponse.json()

        if (!dbStatus.initialized) {
          router.push('/setup')
          return
        }

        // Check activation status
        const activationResponse = await fetch('/api/activation/status')
        const activationStatus = await activationResponse.json()

        if (!activationStatus.isActivated) {
          router.push('/activate')
          return
        }

        // Check if validation expired
        if (activationStatus.validationExpired) {
          router.push('/validation-expired')
          return
        }

        // All checks passed, go to dashboard
        router.push('/dashboard')
      } catch (err) {
        console.error('Setup check error:', err)
        router.push('/dashboard')
      }
    }

    checkSetup()
  }, [router])

  return (
    <main className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
        <p className="text-gray-600">Loading...</p>
      </div>
    </main>
  )
}
