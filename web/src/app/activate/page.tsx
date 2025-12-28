'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'

export default function ActivatePage() {
  const router = useRouter()
  const [licenseKey, setLicenseKey] = useState('')
  const [machineId, setMachineId] = useState('')
  const [isActivating, setIsActivating] = useState(false)
  const [error, setError] = useState('')
  const [showOfflineMode, setShowOfflineMode] = useState(false)
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [isDragging, setIsDragging] = useState(false)

  // Fetch machine ID on mount
  useEffect(() => {
    const fetchMachineId = async () => {
      try {
        const response = await fetch('/api/activation/machine-id')
        const data = await response.json()
        setMachineId(data.machineId)
      } catch (err) {
        console.error('Failed to fetch machine ID:', err)
      }
    }
    fetchMachineId()
  }, [])

  // Validate license key format in real-time
  const validateFormat = (key: string): boolean => {
    const pattern = /^CKSWT-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}$/
    return pattern.test(key)
  }

  // Format license key with hyphens as user types
  const handleLicenseKeyChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    let value = e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, '')

    // Auto-add "CKSWT-" prefix if not present
    if (value.length > 0 && !value.startsWith('CKSWT')) {
      value = 'CKSWT' + value
    }

    // Remove CKSWT for processing
    if (value.startsWith('CKSWT')) {
      value = value.substring(5)
    }

    // Format with hyphens: XXXXX-XXXXX-XXXXX-XXXXX
    const parts = []
    for (let i = 0; i < value.length; i += 5) {
      parts.push(value.substring(i, i + 5))
    }

    const formatted = 'CKSWT-' + parts.join('-')
    setLicenseKey(formatted)
    setError('')
  }

  const handleActivate = async () => {
    if (!validateFormat(licenseKey)) {
      setError('Invalid license key format')
      return
    }

    setIsActivating(true)
    setError('')

    try {
      const response = await fetch('/api/activation/activate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ licenseKey })
      })

      const data = await response.json()

      if (data.success) {
        // Redirect to exercises after successful activation
        router.push('/exercises')
      } else {
        setError(data.error || 'Activation failed')
      }
    } catch (err) {
      setError('Network error. Please check your connection or try Offline Activation.')
    } finally {
      setIsActivating(false)
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && validateFormat(licenseKey) && !isActivating) {
      handleActivate()
    }
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file && (file.name.endsWith('.activation') || file.name.endsWith('.json'))) {
      setSelectedFile(file)
      setError('')
    } else {
      setError('Please select a .activation or .json file')
    }
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
    const file = e.dataTransfer.files?.[0]
    if (file && (file.name.endsWith('.activation') || file.name.endsWith('.json'))) {
      setSelectedFile(file)
      setError('')
    } else {
      setError('Please select a .activation or .json file')
    }
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }

  const handleDragLeave = () => {
    setIsDragging(false)
  }

  const handleOfflineActivate = async () => {
    if (!selectedFile) {
      setError('Please select an activation file')
      return
    }

    setIsActivating(true)
    setError('')

    try {
      // Read the file content
      const fileContent = await selectedFile.text()
      const activationData = JSON.parse(fileContent)

      // Send to offline activation endpoint
      const response = await fetch('/api/activation/activate-offline', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(activationData)
      })

      const data = await response.json()

      if (data.success) {
        router.push('/exercises')
      } else {
        setError(data.error || 'Offline activation failed')
      }
    } catch (err) {
      if (err instanceof SyntaxError) {
        setError('Invalid activation file format')
      } else {
        setError('Failed to process activation file')
      }
    } finally {
      setIsActivating(false)
    }
  }

  const isValidFormat = validateFormat(licenseKey)
  const borderColor = licenseKey.length > 0
    ? (isValidFormat ? 'border-green-500' : 'border-red-500')
    : 'border-gray-300'

  if (showOfflineMode) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
        <div className="max-w-2xl w-full bg-white rounded-lg shadow-md p-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-6">Offline Activation</h1>

          <div className="space-y-6">
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <h2 className="text-lg font-semibold text-blue-900 mb-2">Step 1: Copy Machine ID</h2>
              <p className="text-sm text-blue-800 mb-3">
                Copy this Machine ID and visit the activation website on a device with internet access.
              </p>
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  value={machineId}
                  readOnly
                  className="flex-1 px-4 py-2 border border-blue-300 rounded bg-white font-mono text-sm"
                />
                <button
                  onClick={() => {
                    navigator.clipboard.writeText(machineId)
                    alert('Machine ID copied to clipboard!')
                  }}
                  className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 whitespace-nowrap"
                >
                  Copy ID
                </button>
              </div>
              <p className="text-xs text-blue-700 mt-2">
                Visit: <span className="font-mono">https://activation.cks-weight-room.com/offline</span>
              </p>
            </div>

            <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
              <h2 className="text-lg font-semibold text-gray-900 mb-2">Step 2: Upload Activation File</h2>
              <p className="text-sm text-gray-700 mb-3">
                After obtaining your activation file from the website, upload it here.
              </p>
              <div
                className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
                  isDragging
                    ? 'border-blue-500 bg-blue-50'
                    : selectedFile
                    ? 'border-green-500 bg-green-50'
                    : 'border-gray-300'
                }`}
                onDrop={handleDrop}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
              >
                {selectedFile ? (
                  <div className="text-green-700">
                    <svg className="mx-auto h-12 w-12 mb-2" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                    </svg>
                    <p className="text-sm font-medium">{selectedFile.name}</p>
                    <button
                      onClick={() => setSelectedFile(null)}
                      className="text-xs text-gray-600 hover:text-gray-800 mt-2"
                    >
                      Remove file
                    </button>
                  </div>
                ) : (
                  <div className="text-gray-500">
                    <svg className="mx-auto h-12 w-12 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                    </svg>
                    <p className="text-sm font-medium">Drag and drop activation file here</p>
                    <p className="text-xs text-gray-500 mt-1">or click to browse</p>
                    <input
                      type="file"
                      accept=".activation,.json"
                      onChange={handleFileSelect}
                      className="hidden"
                      id="file-upload"
                    />
                    <label
                      htmlFor="file-upload"
                      className="mt-3 inline-block px-4 py-2 bg-white border border-gray-300 rounded text-sm text-gray-700 hover:bg-gray-50 cursor-pointer"
                    >
                      Browse Files
                    </label>
                  </div>
                )}
              </div>
              <button
                onClick={handleOfflineActivate}
                disabled={!selectedFile || isActivating}
                className={`w-full mt-4 px-4 py-2 rounded font-semibold transition-colors ${
                  !selectedFile || isActivating
                    ? 'bg-gray-400 text-white cursor-not-allowed'
                    : 'bg-green-600 hover:bg-green-700 text-white'
                }`}
              >
                {isActivating ? 'Activating...' : 'Activate Offline'}
              </button>
            </div>

            <button
              onClick={() => setShowOfflineMode(false)}
              className="text-blue-600 hover:text-blue-700 text-sm"
            >
              ← Back to Online Activation
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-md p-8">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Activate CKS Weight Room</h1>
          <p className="text-gray-600">Enter your license key to get started</p>
        </div>

        <div className="space-y-4">
          <div>
            <label htmlFor="licenseKey" className="block text-sm font-medium text-gray-700 mb-2">
              License Key
            </label>
            <input
              id="licenseKey"
              type="text"
              value={licenseKey}
              onChange={handleLicenseKeyChange}
              onKeyPress={handleKeyPress}
              placeholder="CKSWT-XXXXX-XXXXX-XXXXX-XXXXX"
              maxLength={29} // CKSWT-XXXXX-XXXXX-XXXXX-XXXXX = 29 chars
              className={`w-full px-4 py-3 border-2 ${borderColor} rounded-lg font-mono text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors`}
              disabled={isActivating}
            />
            {licenseKey.length > 0 && (
              <p className={`text-xs mt-1 ${isValidFormat ? 'text-green-600' : 'text-red-600'}`}>
                {isValidFormat ? '✓ Valid format' : '✗ Invalid format'}
              </p>
            )}
          </div>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <button
            onClick={handleActivate}
            disabled={!isValidFormat || isActivating}
            className={`w-full py-3 rounded-lg font-semibold text-white transition-colors ${
              !isValidFormat || isActivating
                ? 'bg-gray-400 cursor-not-allowed'
                : 'bg-blue-600 hover:bg-blue-700'
            }`}
          >
            {isActivating ? (
              <span className="flex items-center justify-center gap-2">
                <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Verifying license key...
              </span>
            ) : (
              'Activate'
            )}
          </button>

          <div className="text-center">
            <button
              onClick={() => setShowOfflineMode(true)}
              className="text-sm text-blue-600 hover:text-blue-700"
            >
              Offline Activation
            </button>
          </div>

          {machineId && (
            <div className="bg-gray-50 border border-gray-200 rounded-lg p-3">
              <p className="text-xs text-gray-600 mb-1">Machine ID</p>
              <p className="text-xs font-mono text-gray-800">{machineId}</p>
            </div>
          )}

          <div className="border-t border-gray-200 pt-4">
            <p className="text-xs text-gray-500 text-center">
              Need help? Visit{' '}
              <a href="#" className="text-blue-600 hover:text-blue-700">
                support.cks-weight-room.com
              </a>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
