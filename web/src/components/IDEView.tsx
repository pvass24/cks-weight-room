'use client'

import { useEffect, useState, useRef } from 'react'

interface IDEViewProps {
  exerciseSlug: string
  clusterReady: boolean
  selectedNode: string
}

export default function IDEView({ exerciseSlug, clusterReady, selectedNode }: IDEViewProps) {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const iframeRef = useRef<HTMLIFrameElement>(null)
  const [ideUrl, setIdeUrl] = useState<string>('')

  useEffect(() => {
    if (!clusterReady || !selectedNode) {
      setLoading(true)
      setIdeUrl('')
      return
    }

    // Construct IDE URL with trailing slash so code-server's relative paths resolve correctly
    const url = `/api/ide/${exerciseSlug}/?node=${encodeURIComponent(selectedNode)}`
    console.log('Loading IDE for node:', selectedNode, 'URL:', url)

    setIdeUrl(url)
    setLoading(false)
    setError(null)
  }, [exerciseSlug, clusterReady, selectedNode])

  if (!clusterReady) {
    return (
      <div className="flex-1 bg-gray-900 flex items-center justify-center">
        <div className="text-center text-gray-400">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-500 mx-auto mb-4"></div>
          <p className="text-lg">Waiting for cluster to be ready...</p>
          <p className="text-sm mt-2">The cluster must be provisioned before accessing the IDE</p>
        </div>
      </div>
    )
  }

  if (!selectedNode) {
    return (
      <div className="flex-1 bg-gray-900 flex items-center justify-center">
        <div className="text-center text-gray-400">
          <p className="text-lg">No node selected</p>
          <p className="text-sm mt-2">Please select a node from the dropdown</p>
        </div>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="flex-1 bg-gray-900 flex items-center justify-center">
        <div className="text-center text-gray-400">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-500 mx-auto mb-4"></div>
          <p className="text-lg">Starting IDE on {selectedNode}...</p>
          <p className="text-sm mt-2">This may take a few seconds</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex-1 bg-gray-900 flex items-center justify-center">
        <div className="text-center text-red-400">
          <svg className="h-12 w-12 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p className="text-lg mb-4">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded transition"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex-1 relative bg-gray-900">
      <iframe
        ref={iframeRef}
        src={ideUrl}
        className="absolute inset-0 w-full h-full border-0"
        title={`IDE - ${selectedNode}`}
        allow="clipboard-read; clipboard-write"
      />
    </div>
  )
}
