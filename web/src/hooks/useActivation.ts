import { useEffect, useState } from 'react'
import { useRouter, usePathname } from 'next/navigation'

interface ActivationStatus {
  isActivated: boolean
  licenseKey?: string
  machineId: string
  activatedAt?: string
  expiresAt?: string
  daysRemaining?: number
  inGracePeriod: boolean
  graceDaysLeft?: number
}

export function useActivation(options: { requireActivation?: boolean } = {}) {
  const { requireActivation = false } = options
  const router = useRouter()
  const pathname = usePathname()
  const [status, setStatus] = useState<ActivationStatus | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const checkActivation = async () => {
      try {
        const response = await fetch('/api/activation/status')
        const data: ActivationStatus = await response.json()
        setStatus(data)

        // Redirect to activation page if not activated and activation is required
        if (requireActivation && !data.isActivated && pathname !== '/activate') {
          router.push('/activate')
        }
      } catch (err) {
        console.error('Failed to check activation status:', err)
      } finally {
        setIsLoading(false)
      }
    }

    checkActivation()
  }, [requireActivation, pathname, router])

  return { status, isLoading }
}
