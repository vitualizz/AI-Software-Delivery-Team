import { useState } from 'preact/hooks'
import { SPECIALIST_COLOR } from '@data/artifact-graph'
import type { PipelineFlow } from '@data/pipeline-flow-steps'

const ID_TO_COLOR: Record<string, string> = {
  asdt: 'var(--color-accent)',
  ...Object.fromEntries(
    Object.entries(SPECIALIST_COLOR).map(([id, cssVar]) => [id, `var(${cssVar})`])
  ),
}

interface Props {
  flow: PipelineFlow
  lang: 'en' | 'es'
  labelNext: string
  labelPrev: string
  labelStep: string
  labelOf: string
}

export function FlowWalker({ flow, lang, labelNext, labelPrev, labelStep, labelOf }: Props) {
  const [step, setStep] = useState(0)
  const current = flow.steps[step]
  const total = flow.steps.length
  const isEs = lang === 'es'

  const colorVar = ID_TO_COLOR[current.specialistId] ?? 'var(--color-accent)'

  function handleNext() { if (step < total - 1) setStep(step + 1) }
  function handlePrev() { if (step > 0) setStep(step - 1) }

  return (
    <div class="fw" style={{ '--sc': colorVar } as any}>
      <div class="fw-header">
        <span class="fw-counter">{labelStep} {step + 1} {labelOf} {total}</span>
        <div class="fw-nav">
          <button class="fw-btn" onClick={handlePrev} disabled={step === 0} aria-label={labelPrev}>←</button>
          <button class="fw-btn" onClick={handleNext} disabled={step === total - 1} aria-label={labelNext}>→</button>
        </div>
      </div>
      <code class="fw-command">{current.command}</code>
      <p class="fw-desc">{isEs ? current.descriptionEs : current.descriptionEn}</p>
      <div class="fw-dots" role="tablist">
        {flow.steps.map((_, i) => (
          <button
            key={i}
            class={`fw-dot${i === step ? ' fw-dot--active' : ''}`}
            role="tab"
            aria-selected={i === step}
            aria-label={`${labelStep} ${i + 1}`}
            onClick={() => setStep(i)}
          />
        ))}
      </div>
    </div>
  )
}
