'use client'

import { useEffect, useState } from 'react'

interface TimerProps {
  exerciseSlug: string
  personalBest?: number // in seconds
  onTimeUpdate?: (seconds: number) => void
}

export default function Timer({ exerciseSlug, personalBest, onTimeUpdate }: TimerProps) {
  const [elapsedSeconds, setElapsedSeconds] = useState(0)
  const [isRunning, setIsRunning] = useState(true)

  useEffect(() => {
    if (!isRunning) return

    const interval = setInterval(() => {
      setElapsedSeconds((prev) => {
        const newValue = prev + 1
        if (onTimeUpdate) {
          onTimeUpdate(newValue)
        }
        return newValue
      })
    }, 1000)

    return () => clearInterval(interval)
  }, [isRunning, onTimeUpdate])

  const formatTime = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = seconds % 60

    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
    }
    return `${minutes}:${secs.toString().padStart(2, '0')}`
  }

  const isBestTime = personalBest ? elapsedSeconds < personalBest : false

  return (
    <div className="bg-white rounded-lg shadow-sm p-4 border border-gray-200">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2">
            <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div>
              <div className="text-xs text-gray-500 uppercase tracking-wide">Elapsed Time</div>
              <div className={`text-2xl font-mono font-bold ${isBestTime ? 'text-green-600' : 'text-gray-900'}`}>
                {formatTime(elapsedSeconds)}
                {isBestTime && (
                  <span className="text-xs ml-2 text-green-600">🏆 New best!</span>
                )}
              </div>
            </div>
          </div>

          {personalBest && (
            <div className="ml-6 pl-6 border-l border-gray-200">
              <div className="text-xs text-gray-500 uppercase tracking-wide">Personal Best</div>
              <div className="text-lg font-mono text-gray-600">
                {formatTime(personalBest)}
              </div>
            </div>
          )}
        </div>

        <button
          onClick={() => setIsRunning(!isRunning)}
          className={`px-3 py-1 rounded-md text-sm font-medium ${
            isRunning
              ? 'bg-yellow-100 text-yellow-700 hover:bg-yellow-200'
              : 'bg-green-100 text-green-700 hover:bg-green-200'
          }`}
        >
          {isRunning ? 'Pause' : 'Resume'}
        </button>
      </div>
    </div>
  )
}
