export interface SpecialistItem {
  id: string
  name: string
  desc: string
  command: string
}

export interface PipelineNode {
  id: string
  label: string
}

export interface VsItem {
  chat: string
  asdt: string
}

export interface UIStrings {
  nav: {
    home: string
    docs: string
    github: string
    langPickerLabel: string
    specialists: string
    howItWorks: string
  }
  a11y: {
    skipToContent: string
    themeToggleLabel: string
    langPickerLabel: string
  }
  hero: {
    eyebrow: string
    headline: string
    headlineGrad: string
    sub: string
    cta: string
    secondaryCta: string
    installLabel: string
    installCmd: string
    copyLabel: string
    copiedLabel: string
    copyErrorLabel: string
  }
  specialists: {
    kicker: string
    title: string
    sub: string
    advisorStrip: string
    items: SpecialistItem[]
    orchestrator: SpecialistItem
  }
  terminal: {
    tabs: [string, string, string, string]
  }
  pipeline: {
    title: string
    sub: string
    nodes: PipelineNode[]
    a11yTitle: string
    a11yDesc: string
  }
  recipes: {
    kicker: string
    title: string
    sub: string
    tabs: [string, string, string, string]
    notes: [string, string, string, string]
  }
  vs: {
    kicker: string
    title: string
    sub: string
    chatHead: string
    asdtHead: string
    items: VsItem[]
  }
  ctaBand: {
    title: string
    sub: string
  }
  footer: {
    tagline: string
    githubLabel: string
    docsLabel: string
    licenseLabel: string
    credit: string
  }
  docs: {
    fallbackNotice: string
    fallbackNoticeLink: string
    gettingStarted: string
    specialists: string
    commands: string
    userFlows: string
    onThisPage: string
  }
}
