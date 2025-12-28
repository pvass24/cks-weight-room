'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'

interface AnalyticsData {
  totalPracticeSeconds: number
  scenariosCompleted: number
  totalScenarios: number
  averageCompletionTime: number
  averageScore: number
  personalBestsSet: number
  mockExamsTaken: number
  mockExamsPassed: number
  progressByDomain: DetailedDomain[]
  personalBests: PersonalBest[]
  practiceTimeBreakdown: PracticeTimeBreakdown
}

interface DetailedDomain {
  domain: string
  displayName: string
  weight: number
  completedCount: number
  totalCount: number
  completionPercentage: number
  scenarios: ScenarioProgress[]
}

interface ScenarioProgress {
  slug: string
  title: string
  difficulty: string
  personalBest: number
  attempts: number
  lastPracticed: string
  status: string
}

interface PersonalBest {
  slug: string
  title: string
  domain: string
  domainDisplay: string
  difficulty: string
  personalBest: number
  attempts: number
  lastPracticed: string
}

interface PracticeTimeBreakdown {
  thisWeekSeconds: number
  thisMonthSeconds: number
  allTimeSeconds: number
  averageSessionTime: number
  longestSessionTime: number
}

type SortField = 'time' | 'recent' | 'attempts' | 'domain'

export default function AnalyticsPage() {
  const router = useRouter()
  const [data, setData] = useState<AnalyticsData | null>(null)
  const [loading, setLoading] = useState(true)
  const [sortBy, setSortBy] = useState<SortField>('time')
  const [showExportDialog, setShowExportDialog] = useState(false)
  const [exporting, setExporting] = useState(false)
  const [showResetDialog, setShowResetDialog] = useState(false)
  const [resetting, setResetting] = useState(false)
  const [resetConfirmation, setResetConfirmation] = useState('')
  const [resetStats, setResetStats] = useState<{
    attemptsCount: number
    personalBestsCount: number
    mockExamsCount: number
  } | null>(null)

  useEffect(() => {
    const fetchAnalytics = async () => {
      try {
        const response = await fetch('/api/analytics')
        if (response.ok) {
          const analyticsData = await response.json()
          setData(analyticsData)
        }
      } catch (error) {
        console.error('Failed to fetch analytics:', error)
      } finally {
        setLoading(false)
      }
    }

    fetchAnalytics()
  }, [])

  const formatTime = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = seconds % 60
    if (hours > 0) {
      return `${hours}h ${minutes}m`
    }
    if (minutes > 0) {
      return `${minutes}m ${secs}s`
    }
    return `${secs}s`
  }

  const formatDuration = (seconds: number): string => {
    const minutes = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${minutes}:${secs.toString().padStart(2, '0')}`
  }

  const formatDate = (dateStr: string): string => {
    if (!dateStr) return 'Never'
    try {
      const date = new Date(dateStr)
      const now = new Date()
      const diffMs = now.getTime() - date.getTime()
      const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

      if (diffDays === 0) return 'Today'
      if (diffDays === 1) return 'Yesterday'
      if (diffDays < 7) return `${diffDays} days ago`
      return date.toLocaleDateString()
    } catch {
      return dateStr
    }
  }

  const sortedPersonalBests = data?.personalBests ? [...data.personalBests].sort((a, b) => {
    switch (sortBy) {
      case 'time':
        return a.personalBest - b.personalBest
      case 'recent':
        return new Date(b.lastPracticed || 0).getTime() - new Date(a.lastPracticed || 0).getTime()
      case 'attempts':
        return b.attempts - a.attempts
      case 'domain':
        return a.domainDisplay.localeCompare(b.domainDisplay)
      default:
        return 0
    }
  }) : []

  const exportAsJSON = async () => {
    setExporting(true)
    try {
      const response = await fetch('/api/export')
      if (response.ok) {
        const exportData = await response.json()
        const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `cks-weight-room-progress-${new Date().toISOString().split('T')[0]}.json`
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
        URL.revokeObjectURL(url)
        setShowExportDialog(false)
      }
    } catch (error) {
      console.error('Export failed:', error)
      alert('Failed to export data')
    } finally {
      setExporting(false)
    }
  }

  const exportAsCSV = async () => {
    setExporting(true)
    try {
      const response = await fetch('/api/export')
      if (response.ok) {
        const exportData = await response.json()

        // Create CSV content
        const headers = ['Attempt ID', 'Scenario ID', 'Scenario Name', 'Timestamp', 'Completion Time (MM:SS)', 'Score (%)', 'Status', 'Feedback']
        const rows = exportData.attempts.map((attempt: any) => {
          const minutes = Math.floor(attempt.completion_time_seconds / 60)
          const seconds = attempt.completion_time_seconds % 60
          const time = `${minutes}:${seconds.toString().padStart(2, '0')}`
          const score = Math.round(attempt.score * 100)
          return [
            attempt.attempt_id,
            attempt.scenario_id,
            `"${attempt.scenario_name}"`,
            attempt.timestamp,
            time,
            score,
            attempt.status,
            `"${attempt.feedback || ''}"`
          ].join(',')
        })

        const csv = [headers.join(','), ...rows].join('\n')
        const blob = new Blob([csv], { type: 'text/csv' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `cks-weight-room-progress-${new Date().toISOString().split('T')[0]}.csv`
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
        URL.revokeObjectURL(url)
        setShowExportDialog(false)
      }
    } catch (error) {
      console.error('Export failed:', error)
      alert('Failed to export data')
    } finally {
      setExporting(false)
    }
  }

  const handleResetClick = async () => {
    try {
      const response = await fetch('/api/reset/stats')
      if (response.ok) {
        const stats = await response.json()
        setResetStats(stats)
        setShowResetDialog(true)
        setResetConfirmation('')
      }
    } catch (error) {
      console.error('Failed to fetch reset stats:', error)
      alert('Failed to load reset statistics')
    }
  }

  const handleResetConfirm = async () => {
    if (resetConfirmation !== 'DELETE') {
      return
    }

    setResetting(true)
    try {
      const response = await fetch('/api/reset', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ confirmation: resetConfirmation }),
      })

      if (response.ok) {
        const result = await response.json()
        if (result.success) {
          setShowResetDialog(false)
          alert(result.message)
          router.push('/exercises')
        } else {
          alert(result.message)
        }
      } else {
        alert('Failed to reset progress')
      }
    } catch (error) {
      console.error('Reset failed:', error)
      alert('Failed to reset progress')
    } finally {
      setResetting(false)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading analytics...</p>
        </div>
      </div>
    )
  }

  if (!data) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-gray-600">No analytics data available</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <button
              onClick={() => router.push('/')}
              className="flex items-center text-gray-600 hover:text-gray-900"
            >
              <svg className="w-5 h-5 mr-2 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
              Back to Home
            </button>
            <button
              onClick={() => setShowExportDialog(true)}
              className="bg-blue-600 hover:bg-blue-700 text-white font-semibold py-2 px-6 rounded-lg transition-colors flex items-center gap-2"
            >
              <svg className="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              Export Data
            </button>
          </div>
          <h1 className="text-4xl font-bold text-gray-900">Analytics Dashboard</h1>
          <p className="text-gray-600 mt-2">Comprehensive view of your practice progress</p>
        </div>

        {/* Top-Level Statistics */}
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 mb-8">
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-600">Total Practice Time</div>
            <div className="text-2xl font-bold text-blue-600">{formatTime(data.totalPracticeSeconds)}</div>
          </div>

          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-600">Scenarios Completed</div>
            <div className="text-2xl font-bold text-green-600">
              {data.scenariosCompleted}/{data.totalScenarios}
            </div>
            <div className="text-xs text-gray-500">
              {data.totalScenarios > 0 ? Math.round((data.scenariosCompleted / data.totalScenarios) * 100) : 0}%
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-600">Avg Completion Time</div>
            <div className="text-2xl font-bold text-purple-600">
              {data.averageCompletionTime > 0 ? formatDuration(data.averageCompletionTime) : '—'}
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-600">Average Score</div>
            <div className="text-2xl font-bold text-indigo-600">
              {data.averageScore > 0 ? Math.round(data.averageScore) : '—'}%
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-600">Personal Bests</div>
            <div className="text-2xl font-bold text-yellow-600">{data.personalBestsSet}</div>
          </div>

          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-600">Mock Exams</div>
            <div className="text-2xl font-bold text-orange-600">{data.mockExamsTaken}</div>
            <div className="text-xs text-gray-500">{data.mockExamsPassed} passed</div>
          </div>
        </div>

        {/* Practice Time Breakdown */}
        <div className="bg-white rounded-lg shadow p-6 mb-8">
          <h2 className="text-xl font-bold text-gray-900 mb-4">Practice Time Breakdown</h2>
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            <div>
              <div className="text-sm text-gray-600">This Week</div>
              <div className="text-lg font-semibold text-gray-900">
                {formatTime(data.practiceTimeBreakdown.thisWeekSeconds)}
              </div>
            </div>
            <div>
              <div className="text-sm text-gray-600">This Month</div>
              <div className="text-lg font-semibold text-gray-900">
                {formatTime(data.practiceTimeBreakdown.thisMonthSeconds)}
              </div>
            </div>
            <div>
              <div className="text-sm text-gray-600">All Time</div>
              <div className="text-lg font-semibold text-gray-900">
                {formatTime(data.practiceTimeBreakdown.allTimeSeconds)}
              </div>
            </div>
            <div>
              <div className="text-sm text-gray-600">Avg Session</div>
              <div className="text-lg font-semibold text-gray-900">
                {data.practiceTimeBreakdown.averageSessionTime > 0
                  ? formatTime(data.practiceTimeBreakdown.averageSessionTime)
                  : '—'}
              </div>
            </div>
            <div>
              <div className="text-sm text-gray-600">Longest Session</div>
              <div className="text-lg font-semibold text-gray-900">
                {data.practiceTimeBreakdown.longestSessionTime > 0
                  ? formatTime(data.practiceTimeBreakdown.longestSessionTime)
                  : '—'}
              </div>
            </div>
          </div>
        </div>

        {/* Progress by Domain */}
        <div className="bg-white rounded-lg shadow p-6 mb-8">
          <h2 className="text-xl font-bold text-gray-900 mb-4">Progress by Domain</h2>
          <div className="space-y-6">
            {data.progressByDomain.map((domain) => (
              <div key={domain.domain}>
                <div className="flex items-center justify-between mb-2">
                  <div>
                    <div className="font-semibold text-gray-900">{domain.displayName}</div>
                    <div className="text-xs text-gray-500">{domain.weight}% of exam</div>
                  </div>
                  <div className="text-right">
                    <div className="font-bold text-gray-900">
                      {domain.completedCount}/{domain.totalCount}
                    </div>
                    <div className="text-xs text-gray-600">
                      {Math.round(domain.completionPercentage)}%
                    </div>
                  </div>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2 mb-3">
                  <div
                    className="bg-blue-600 h-2 rounded-full transition-all"
                    style={{ width: `${domain.completionPercentage}%` }}
                  ></div>
                </div>
                <div className="ml-4 space-y-1">
                  {domain.scenarios.map((scenario) => (
                    <div key={scenario.slug} className="flex items-center justify-between text-sm py-1">
                      <div className="flex-1">
                        <span className="text-gray-700">{scenario.title}</span>
                        <span className={`ml-2 text-xs px-2 py-0.5 rounded ${
                          scenario.difficulty === 'easy' ? 'bg-green-100 text-green-700' :
                          scenario.difficulty === 'medium' ? 'bg-yellow-100 text-yellow-700' :
                          'bg-red-100 text-red-700'
                        }`}>
                          {scenario.difficulty}
                        </span>
                      </div>
                      <div className="text-right text-xs text-gray-600">
                        {scenario.status === 'not-started' ? (
                          <span className="text-gray-400">Not Started</span>
                        ) : (
                          <>
                            <span className="font-medium">Best: {formatDuration(scenario.personalBest)}</span>
                            <span className="ml-2">({scenario.attempts} attempts)</span>
                          </>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Personal Bests Table */}
        {data.personalBests.length > 0 && (
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-bold text-gray-900">Personal Bests</h2>
              <div className="flex gap-2">
                <button
                  onClick={() => setSortBy('time')}
                  className={`px-3 py-1 text-sm rounded ${
                    sortBy === 'time' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'
                  }`}
                >
                  Fastest
                </button>
                <button
                  onClick={() => setSortBy('recent')}
                  className={`px-3 py-1 text-sm rounded ${
                    sortBy === 'recent' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'
                  }`}
                >
                  Recent
                </button>
                <button
                  onClick={() => setSortBy('attempts')}
                  className={`px-3 py-1 text-sm rounded ${
                    sortBy === 'attempts' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'
                  }`}
                >
                  Most Attempts
                </button>
                <button
                  onClick={() => setSortBy('domain')}
                  className={`px-3 py-1 text-sm rounded ${
                    sortBy === 'domain' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'
                  }`}
                >
                  Domain
                </button>
              </div>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-gray-200">
                    <th className="text-left py-3 px-4 font-semibold text-gray-700">Scenario</th>
                    <th className="text-left py-3 px-4 font-semibold text-gray-700">Domain</th>
                    <th className="text-left py-3 px-4 font-semibold text-gray-700">Difficulty</th>
                    <th className="text-left py-3 px-4 font-semibold text-gray-700">Best Time</th>
                    <th className="text-left py-3 px-4 font-semibold text-gray-700">Attempts</th>
                    <th className="text-left py-3 px-4 font-semibold text-gray-700">Last Practiced</th>
                  </tr>
                </thead>
                <tbody>
                  {sortedPersonalBests.map((pb) => (
                    <tr key={pb.slug} className="border-b border-gray-100 hover:bg-gray-50">
                      <td className="py-3 px-4 text-gray-900">{pb.title}</td>
                      <td className="py-3 px-4 text-gray-600 text-sm">{pb.domainDisplay}</td>
                      <td className="py-3 px-4">
                        <span className={`text-xs px-2 py-1 rounded ${
                          pb.difficulty === 'easy' ? 'bg-green-100 text-green-700' :
                          pb.difficulty === 'medium' ? 'bg-yellow-100 text-yellow-700' :
                          'bg-red-100 text-red-700'
                        }`}>
                          {pb.difficulty}
                        </span>
                      </td>
                      <td className="py-3 px-4 font-semibold text-blue-600">
                        {formatDuration(pb.personalBest)}
                      </td>
                      <td className="py-3 px-4 text-gray-600">{pb.attempts}</td>
                      <td className="py-3 px-4 text-gray-600 text-sm">{formatDate(pb.lastPracticed)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* Danger Zone */}
        <div className="bg-red-50 border border-red-200 rounded-lg p-6 mt-8">
          <h2 className="text-xl font-bold text-red-900 mb-2">Danger Zone</h2>
          <p className="text-red-700 mb-4">
            Permanently delete all your progress data. This action cannot be undone.
          </p>
          <button
            onClick={handleResetClick}
            className="bg-red-600 hover:bg-red-700 text-white font-semibold py-2 px-6 rounded-lg transition-colors"
          >
            Reset All Progress
          </button>
        </div>

        {/* Reset Confirmation Dialog */}
        {showResetDialog && resetStats && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg shadow-xl p-6 max-w-md w-full mx-4">
              <h2 className="text-2xl font-bold text-red-600 mb-4">⚠️ Reset All Progress</h2>
              <p className="text-gray-700 mb-4">
                Are you sure you want to reset <strong>ALL</strong> progress data? This will permanently delete:
              </p>
              <ul className="text-gray-700 mb-4 space-y-2">
                <li>• <strong>{resetStats.attemptsCount}</strong> scenario attempts</li>
                <li>• <strong>{resetStats.personalBestsCount}</strong> personal best records</li>
                <li>• <strong>{resetStats.mockExamsCount}</strong> mock exam results</li>
                <li>• All progress statistics</li>
              </ul>
              <div className="bg-yellow-50 border border-yellow-200 rounded p-3 mb-4">
                <p className="text-sm text-yellow-800">
                  <strong>This action CANNOT be undone.</strong> Consider exporting your data first.
                </p>
              </div>
              <div className="mb-4">
                <label className="block text-sm font-semibold text-gray-700 mb-2">
                  Type <code className="bg-gray-100 px-2 py-1 rounded text-red-600">DELETE</code> to confirm:
                </label>
                <input
                  type="text"
                  value={resetConfirmation}
                  onChange={(e) => setResetConfirmation(e.target.value)}
                  placeholder="Type DELETE"
                  disabled={resetting}
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500"
                />
              </div>
              <div className="flex gap-3">
                <button
                  onClick={() => setShowResetDialog(false)}
                  disabled={resetting}
                  className="flex-1 bg-gray-200 hover:bg-gray-300 disabled:bg-gray-100 text-gray-700 font-semibold py-2 px-4 rounded-lg transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={handleResetConfirm}
                  disabled={resetting || resetConfirmation !== 'DELETE'}
                  className="flex-1 bg-red-600 hover:bg-red-700 disabled:bg-gray-400 text-white font-semibold py-2 px-4 rounded-lg transition-colors"
                >
                  {resetting ? 'Resetting...' : 'Reset All Progress'}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Export Dialog */}
        {showExportDialog && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg shadow-xl p-6 max-w-md w-full mx-4">
              <h2 className="text-2xl font-bold text-gray-900 mb-4">Export Progress Data</h2>
              <p className="text-gray-600 mb-6">
                Choose a format to export your practice data. Your data is exported locally and never sent to external servers.
              </p>

              <div className="space-y-4 mb-6">
                <div className="border border-gray-200 rounded-lg p-4">
                  <h3 className="font-semibold text-gray-900 mb-2">What will be exported:</h3>
                  <ul className="text-sm text-gray-600 space-y-1">
                    <li>• All scenario attempts with timestamps and scores</li>
                    <li>• Personal best times for each scenario</li>
                    <li>• Mock exam results</li>
                    <li>• Overall practice statistics</li>
                  </ul>
                </div>
              </div>

              <div className="flex gap-3">
                <button
                  onClick={exportAsJSON}
                  disabled={exporting}
                  className="flex-1 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-semibold py-3 px-4 rounded-lg transition-colors"
                >
                  {exporting ? 'Exporting...' : 'Export as JSON'}
                </button>
                <button
                  onClick={exportAsCSV}
                  disabled={exporting}
                  className="flex-1 bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white font-semibold py-3 px-4 rounded-lg transition-colors"
                >
                  {exporting ? 'Exporting...' : 'Export as CSV'}
                </button>
              </div>

              <button
                onClick={() => setShowExportDialog(false)}
                disabled={exporting}
                className="w-full mt-3 bg-gray-200 hover:bg-gray-300 disabled:bg-gray-100 text-gray-700 font-semibold py-2 px-4 rounded-lg transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
