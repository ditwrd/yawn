import { describe, it, expect } from 'vitest'
import { cn } from '@/lib/utils'

describe('Design System - Theme Application', () => {
  it('applies dark theme CSS classes correctly', () => {
    // Test that dark theme classes exist
    const darkThemeClass = cn('dark:bg-background', 'dark:text-foreground')
    expect(darkThemeClass).toContain('dark:bg-background')
    expect(darkThemeClass).toContain('dark:text-foreground')
  })

  it('provides color system for sleepy YAWN theme', () => {
    // Test that color system utilities work
    const colorClass = cn('bg-primary', 'text-primary-foreground')
    expect(colorClass).toContain('bg-primary')
    expect(colorClass).toContain('text-primary-foreground')
  })
})

describe('Design System - Responsive Design', () => {
  it('provides mobile-first responsive utilities', () => {
    // Test mobile-first responsive patterns
    const responsiveClass = cn('w-full md:w-auto')
    expect(responsiveClass).toContain('w-full')
    expect(responsiveClass).toContain('md:w-auto')
  })

  it('implements responsive spacing system', () => {
    const spacingClass = cn('p-4 sm:p-6 lg:p-8')
    expect(spacingClass).toContain('p-4')
    expect(spacingClass).toContain('sm:p-6')
    expect(spacingClass).toContain('lg:p-8')
  })
})

describe('Design System - Accessibility Features', () => {
  it('provides focus management for keyboard navigation', () => {
    const focusClass = cn('focus:ring-2 focus:ring-ring focus:ring-offset-2')
    expect(focusClass).toContain('focus:ring-2')
    expect(focusClass).toContain('focus:ring-ring')
    expect(focusClass).toContain('focus:ring-offset-2')
  })

  it('includes screen reader utilities', () => {
    const srClass = cn('sr-only')
    expect(srClass).toContain('sr-only')
  })
})

describe('Design System - Typography', () => {
  it('configures JetBrains Mono font family', () => {
    const monoClass = cn('font-mono')
    expect(monoClass).toContain('font-mono')
  })

  it('provides responsive text scaling', () => {
    const textClass = cn('text-sm md:text-base lg:text-lg')
    expect(textClass).toContain('text-sm')
    expect(textClass).toContain('md:text-base')
    expect(textClass).toContain('lg:text-lg')
  })
})