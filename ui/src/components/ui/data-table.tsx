import * as React from "react"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "./table"
import { Button } from "./button"
import { Input } from "./input"
import { Badge } from "./badge"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "./dropdown-menu"
import { cn } from "@/lib/utils"
import {
  ChevronDown,
  ChevronUp,
  MoreHorizontal,
  ArrowUpDown,
  Search,
  Filter
} from "lucide-react"

type SortDirection = "asc" | "desc" | null

interface Column<T> {
  key: keyof T
  title: string
  sortable?: boolean
  filterable?: boolean
  render?: (value: any, row: T) => React.ReactNode
  width?: string
}

interface DataTableProps<T> {
  data: T[]
  columns: Column<T>[]
  className?: string
  searchable?: boolean
  searchPlaceholder?: string
  emptyMessage?: string
  onRowClick?: (row: T) => void
  actions?: (row: T) => React.ReactNode
  selection?: {
    selected: string[]
    onSelectionChange: (selected: string[]) => void
    getRowId: (row: T) => string
  }
}

export function DataTable<T>({
  data,
  columns,
  className,
  searchable = true,
  searchPlaceholder = "Search...",
  emptyMessage = "No data available",
  onRowClick,
  actions,
  selection,
}: DataTableProps<T>) {
  const [sortConfig, setSortConfig] = React.useState<{
    key: keyof T
    direction: SortDirection
  } | null>(null)
  const [filters, setFilters] = React.useState<Record<string, string>>({})
  const [searchQuery, setSearchQuery] = React.useState("")
  const [filteredData, setFilteredData] = React.useState<T[]>(data)

  // Apply filters and search
  React.useEffect(() => {
    let result = [...data]

    // Apply search
    if (searchQuery) {
      result = result.filter((row) =>
        columns.some((column) => {
          const value = row[column.key]
          return String(value).toLowerCase().includes(searchQuery.toLowerCase())
        })
      )
    }

    // Apply column filters
    Object.entries(filters).forEach(([key, filterValue]) => {
      if (filterValue) {
        result = result.filter((row) => {
          const value = row[key as keyof T]
          return String(value).toLowerCase().includes(filterValue.toLowerCase())
        })
      }
    })

    // Apply sorting
    if (sortConfig && sortConfig.direction) {
      result.sort((a, b) => {
        const aValue = a[sortConfig.key]
        const bValue = b[sortConfig.key]

        if (aValue === null || aValue === undefined) return 1
        if (bValue === null || bValue === undefined) return -1

        const comparison = String(aValue).localeCompare(String(bValue))
        return sortConfig.direction === "asc" ? comparison : -comparison
      })
    }

    setFilteredData(result)
  }, [data, searchQuery, filters, sortConfig, columns])

  const handleSort = (key: keyof T) => {
    setSortConfig((current) => {
      if (!current || current.key !== key) {
        return { key, direction: "asc" }
      }
      if (current.direction === "asc") {
        return { key, direction: "desc" }
      }
      return null
    })
  }

  const handleFilter = (key: string, value: string) => {
    setFilters((current) => ({
      ...current,
      [key]: value,
    }))
  }

  const toggleRowSelection = (rowId: string) => {
    if (!selection) return

    const newSelection = selection.selected.includes(rowId)
      ? selection.selected.filter((id) => id !== rowId)
      : [...selection.selected, rowId]

    selection.onSelectionChange(newSelection)
  }

  const toggleAllRows = () => {
    if (!selection) return

    const allRowIds = filteredData.map(selection.getRowId)
    const allSelected = allRowIds.every((id) => selection.selected.includes(id))

    selection.onSelectionChange(
      allSelected ? [] : allRowIds
    )
  }

  const getSortIcon = (columnKey: keyof T) => {
    if (sortConfig?.key !== columnKey) {
      return <ArrowUpDown className="ml-2 h-4 w-4" />
    }
    return sortConfig.direction === "asc" ? (
      <ChevronUp className="ml-2 h-4 w-4" />
    ) : (
      <ChevronDown className="ml-2 h-4 w-4" />
    )
  }

  return (
    <div className={cn("w-full space-y-4", className)}>
      {/* Search and Filters */}
      {searchable && (
        <div className="flex items-center gap-4">
          <div className="relative flex-1 max-w-sm">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder={searchPlaceholder}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>

          {columns.filter((col) => col.filterable).length > 0 && (
            <div className="flex items-center gap-2">
              <Filter className="h-4 w-4 text-muted-foreground" />
              {columns
                .filter((col) => col.filterable)
                .map((column) => (
                  <Input
                    key={String(column.key)}
                    placeholder={`Filter ${column.title}`}
                    value={filters[String(column.key)] || ""}
                    onChange={(e) =>
                      handleFilter(String(column.key), e.target.value)
                    }
                    className="w-32"
                  />
                ))}
            </div>
          )}
        </div>
      )}

      {/* Table */}
      <div className="rounded-md border border-border">
        <Table>
          <TableHeader>
            <TableRow>
              {selection && (
                <TableHead className="w-12">
                  <input
                    type="checkbox"
                    checked={
                      filteredData.length > 0 &&
                      filteredData.every((row) =>
                        selection.selected.includes(selection.getRowId(row))
                      )
                    }
                    onChange={toggleAllRows}
                    className="rounded border-gray-300"
                  />
                </TableHead>
              )}
              {columns.map((column) => (
                <TableHead
                  key={String(column.key)}
                  style={{ width: column.width }}
                  className={cn(
                    "font-medium",
                    column.sortable && "cursor-pointer hover:bg-muted/50"
                  )}
                  onClick={() => column.sortable && handleSort(column.key)}
                >
                  <div className="flex items-center">
                    {column.title}
                    {column.sortable && getSortIcon(column.key)}
                  </div>
                </TableHead>
              ))}
              {actions && <TableHead className="w-12" />}
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredData.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={
                    columns.length + (selection ? 1 : 0) + (actions ? 1 : 0)
                  }
                  className="h-24 text-center text-muted-foreground"
                >
                  {emptyMessage}
                </TableCell>
              </TableRow>
            ) : (
              filteredData.map((row, index) => {
                const rowId = selection ? selection.getRowId(row) : String(index)
                const isSelected = selection ? selection.selected.includes(rowId) : false

                return (
                  <TableRow
                    key={rowId}
                    className={cn(
                      onRowClick && "cursor-pointer hover:bg-muted/50",
                      isSelected && "bg-muted/30"
                    )}
                    onClick={() => onRowClick?.(row)}
                  >
                    {selection && (
                      <TableCell>
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => toggleRowSelection(rowId)}
                          onClick={(e) => e.stopPropagation()}
                          className="rounded border-gray-300"
                        />
                      </TableCell>
                    )}
                    {columns.map((column) => (
                      <TableCell key={String(column.key)}>
                        {column.render
                          ? column.render(row[column.key], row)
                          : String(row[column.key] ?? "")}
                      </TableCell>
                    ))}
                    {actions && (
                      <TableCell>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button
                              variant="ghost"
                              className="h-8 w-8 p-0"
                              onClick={(e) => e.stopPropagation()}
                            >
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            {actions(row)}
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    )}
                  </TableRow>
                )
              })
            )}
          </TableBody>
        </Table>
      </div>

      {/* Results count */}
      <div className="text-sm text-muted-foreground">
        {filteredData.length} of {data.length} results
      </div>
    </div>
  )
}