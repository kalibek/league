import { useRef, useState } from 'react'
import { Button } from '../Button/Button'

interface ImportResult {
  imported: number
  skipped: number
  errors: string[]
}

interface CSVImportProps {
  onImport: (file: File) => Promise<ImportResult | null>
  loading?: boolean
}

export function CSVImport({ onImport, loading = false }: CSVImportProps) {
  const [file, setFile] = useState<File | null>(null)
  const [preview, setPreview] = useState<string[][]>([])
  const [result, setResult] = useState<ImportResult | null>(null)
  const [dragOver, setDragOver] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  const readCSV = (f: File) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      const text = e.target?.result as string
      const rows = text
        .split('\n')
        .slice(0, 6)
        .map((row) => row.split(',').map((cell) => cell.trim()))
        .filter((row) => row.some((c) => c !== ''))
      setPreview(rows)
    }
    reader.readAsText(f)
  }

  const handleFile = (f: File) => {
    setFile(f)
    setResult(null)
    readCSV(f)
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)
    const f = e.dataTransfer.files[0]
    if (f && f.name.endsWith('.csv')) handleFile(f)
  }

  const handleSubmit = async () => {
    if (!file) return
    const r = await onImport(file)
    setResult(r)
  }

  return (
    <div className="flex flex-col gap-4">
      {/* Drop zone */}
      <div
        onDragOver={(e) => { e.preventDefault(); setDragOver(true) }}
        onDragLeave={() => setDragOver(false)}
        onDrop={handleDrop}
        onClick={() => inputRef.current?.click()}
        className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors ${
          dragOver ? 'border-blue-400 bg-blue-50' : 'border-gray-300 hover:border-gray-400'
        }`}
      >
        <input
          ref={inputRef}
          type="file"
          accept=".csv"
          className="hidden"
          onChange={(e) => {
            const f = e.target.files?.[0]
            if (f) handleFile(f)
          }}
        />
        {file ? (
          <p className="text-sm text-gray-700 font-medium">{file.name}</p>
        ) : (
          <>
            <p className="text-gray-500">Drag and drop a CSV file here, or click to select</p>
            <p className="text-xs text-gray-400 mt-1">
              Columns: first_name, last_name, email, initial_rating (optional)
            </p>
          </>
        )}
      </div>

      {/* Preview */}
      {preview.length > 0 && (
        <div className="overflow-x-auto rounded-md border border-gray-200">
          <table className="w-full text-xs">
            <thead className="bg-gray-50">
              <tr>
                {preview[0].map((cell, i) => (
                  <th key={i} className="px-3 py-2 text-left text-gray-600 font-medium">
                    {cell}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {preview.slice(1).map((row, ri) => (
                <tr key={ri}>
                  {row.map((cell, ci) => (
                    <td key={ci} className="px-3 py-2 text-gray-700">
                      {cell}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Submit */}
      {file && (
        <Button onClick={handleSubmit} loading={loading} disabled={!file}>
          Import Players
        </Button>
      )}

      {/* Result */}
      {result && (
        <div className="rounded-md border border-gray-200 p-4 bg-gray-50 text-sm">
          <p className="font-medium text-gray-800">
            Imported: {result.imported} &nbsp;|&nbsp; Skipped: {result.skipped} (duplicates)
            {(result.errors ?? []).length > 0 && (
              <span className="text-red-600"> &nbsp;|&nbsp; Errors: {result.errors.length}</span>
            )}
          </p>
          {(result.errors ?? []).length > 0 && (
            <ul className="mt-2 space-y-1 text-red-600 text-xs">
              {result.errors.map((e, i) => (
                <li key={i}>{e}</li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  )
}
