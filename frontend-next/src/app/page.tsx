import type { Metadata } from "next";
import Link from "next/link";
import {
  Lightbulb,
  Brain,
  CheckCircle,
  FileText,
  Globe,
  Compass,
  Layers,
  Zap,
  Rocket,
  Code,
  CalendarDays,
  Target,
  ArrowRight,
  Check,
} from "lucide-react";
import { TIER_FEATURES } from "@/lib/constants";

export const metadata: Metadata = {
  title: "GoPlan — AI Strategy Consultant",
  description:
    "Turn any idea into a structured strategic plan. AI-powered strategy consulting for founders, teams, and enterprises.",
  openGraph: {
    title: "GoPlan — AI Strategy Consultant",
    description:
      "Turn any idea into a structured strategic plan. AI-powered strategy consulting for founders, teams, and enterprises.",
    type: "website",
  },
};

const steps = [
  {
    number: 1,
    icon: Lightbulb,
    title: "Describe your idea",
    description:
      "Enter your concept, initiative, or business idea in plain language. No templates, no forms — just tell us what you're planning.",
  },
  {
    number: 2,
    icon: Brain,
    title: "AI generates your strategy",
    description:
      "Our AI analyzes your input and produces a full strategic framework with executive briefing, market context, and phased execution plan.",
  },
  {
    number: 3,
    icon: CheckCircle,
    title: "Refine and execute",
    description:
      "Iterate on individual sections, regenerate with deeper analysis, export to PDF, and share with your team.",
  },
];

const frameworkSections = [
  {
    icon: FileText,
    title: "Executive Brief",
    description:
      "Quick summary of strategy, objectives, and expected outcomes",
  },
  {
    icon: Globe,
    title: "Strategic Context",
    description:
      "Industry analysis, market conditions, and competitive landscape",
  },
  {
    icon: Compass,
    title: "Recommended Approach",
    description:
      "Core strategy with rationale, key pillars, and risk mitigation",
  },
  {
    icon: Layers,
    title: "Phased Execution Plan",
    description:
      "Concrete phases with milestones, actions, and success criteria",
  },
  {
    icon: Zap,
    title: "Immediate Action Plan",
    description:
      "First 30 days: quick wins, critical path, and resource needs",
  },
];

const useCases = [
  {
    icon: Rocket,
    title: "Launching a business",
    example: '"I want to open a premium coffee roastery in Austin, TX"',
    output:
      "Market analysis, competitive positioning, launch phases, and a 90-day action plan with milestones.",
  },
  {
    icon: Code,
    title: "Building a SaaS product",
    example: '"B2B invoice automation tool for small accounting firms"',
    output:
      "Go-to-market strategy, pricing model, technical roadmap, and user acquisition plan.",
  },
  {
    icon: CalendarDays,
    title: "Planning an event",
    example: '"500-person tech conference in Q3 with sponsors and speakers"',
    output:
      "Venue logistics, sponsorship tiers, marketing timeline, and week-by-week execution schedule.",
  },
  {
    icon: Target,
    title: "Strategic initiatives",
    example: '"Expand our fintech platform into the European market"',
    output:
      "Regulatory considerations, localization strategy, phased rollout, and risk assessment.",
  },
];

const pricingTiers = [
  {
    name: "Free",
    price: "$0",
    period: "/mo",
    description: "For individuals exploring strategic planning",
    features: TIER_FEATURES.free,
    cta: "Get Started",
    href: "/register",
    highlighted: false,
  },
  {
    name: "Pro",
    price: "$49",
    period: "/mo",
    description: "For founders and teams who plan seriously",
    features: TIER_FEATURES.pro,
    cta: "Upgrade to Pro",
    href: "/register",
    highlighted: true,
    badge: "Most Popular",
  },
  {
    name: "Pro+",
    price: "$99",
    period: "/mo",
    description: "For enterprises and power users",
    features: TIER_FEATURES.pro_plus,
    cta: "Go Pro+",
    href: "/register",
    highlighted: false,
  },
];

export default function Home() {
  return (
    <main>
      {/* ===== HERO ===== */}
      <section className="relative overflow-hidden bg-navy-950 px-6 py-24 sm:py-32 lg:py-40">
        {/* Subtle grid pattern */}
        <div
          className="pointer-events-none absolute inset-0 opacity-[0.03]"
          style={{
            backgroundImage:
              "linear-gradient(rgba(255,255,255,.1) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,.1) 1px, transparent 1px)",
            backgroundSize: "64px 64px",
          }}
        />
        {/* Gradient glow */}
        <div className="pointer-events-none absolute left-1/2 top-0 -translate-x-1/2 h-[600px] w-[800px] rounded-full bg-brand/10 blur-3xl" />

        <div className="relative mx-auto max-w-4xl text-center">
          <h1 className="text-4xl font-bold tracking-tight text-white md:text-6xl">
            Turn any idea into a structured strategic plan
          </h1>
          <p className="mx-auto mt-6 max-w-2xl text-lg leading-8 text-navy-300">
            GoPlan is your AI strategy consultant. Get executive-level strategic
            analysis in seconds, not weeks.
          </p>
          <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              href="/register"
              className="inline-flex items-center gap-2 rounded-lg bg-brand px-8 py-3.5 text-base font-semibold text-white shadow-lg shadow-brand/25 transition-all hover:bg-brand-dark hover:shadow-xl hover:shadow-brand/30"
            >
              Start Planning
              <ArrowRight className="h-4 w-4" />
            </Link>
            <a
              href="#how-it-works"
              className="inline-flex items-center gap-2 rounded-lg border border-navy-600 px-8 py-3.5 text-base font-semibold text-navy-200 transition-colors hover:border-navy-400 hover:text-white"
            >
              See how it works
            </a>
          </div>
        </div>
      </section>

      {/* ===== HOW IT WORKS ===== */}
      <section id="how-it-works" className="bg-white px-6 py-24 sm:py-32">
        <div className="mx-auto max-w-5xl">
          <div className="text-center">
            <h2 className="text-3xl font-bold tracking-tight text-navy-900 sm:text-4xl">
              How it works
            </h2>
            <p className="mt-4 text-lg text-navy-500">
              From idea to actionable strategy in three simple steps.
            </p>
          </div>

          <div className="mt-16 grid grid-cols-1 gap-12 md:grid-cols-3 md:gap-8">
            {steps.map((step) => (
              <div key={step.number} className="text-center">
                <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-navy-50">
                  <step.icon className="h-7 w-7 text-brand" />
                </div>
                <div className="mt-2 inline-flex h-7 w-7 items-center justify-center rounded-full bg-navy-900 text-xs font-bold text-white">
                  {step.number}
                </div>
                <h3 className="mt-4 text-xl font-semibold text-navy-900">
                  {step.title}
                </h3>
                <p className="mt-3 text-base leading-7 text-navy-500">
                  {step.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ===== FRAMEWORK PREVIEW ===== */}
      <section className="bg-navy-50 px-6 py-24 sm:py-32">
        <div className="mx-auto max-w-5xl">
          <div className="text-center">
            <h2 className="text-3xl font-bold tracking-tight text-navy-900 sm:text-4xl">
              A structured framework for every strategic decision
            </h2>
            <p className="mt-4 text-lg text-navy-500">
              Every plan is built on five comprehensive sections designed by
              strategy professionals.
            </p>
          </div>

          <div className="mt-16 grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {frameworkSections.map((section) => (
              <div
                key={section.title}
                className="group rounded-xl border border-navy-200 bg-white p-6 transition-all hover:border-brand/40 hover:shadow-lg hover:shadow-brand/5"
              >
                <div className="flex h-11 w-11 items-center justify-center rounded-lg bg-navy-50 transition-colors group-hover:bg-brand/10">
                  <section.icon className="h-5 w-5 text-navy-700 transition-colors group-hover:text-brand" />
                </div>
                <h3 className="mt-4 text-lg font-semibold text-navy-900">
                  {section.title}
                </h3>
                <p className="mt-2 text-sm leading-6 text-navy-500">
                  {section.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ===== USE CASES ===== */}
      <section className="bg-white px-6 py-24 sm:py-32">
        <div className="mx-auto max-w-5xl">
          <div className="text-center">
            <h2 className="text-3xl font-bold tracking-tight text-navy-900 sm:text-4xl">
              Built for any strategic initiative
            </h2>
            <p className="mt-4 text-lg text-navy-500">
              Whether you are launching a startup or planning a corporate
              expansion, GoPlan adapts to your domain.
            </p>
          </div>

          <div className="mt-16 grid grid-cols-1 gap-8 sm:grid-cols-2">
            {useCases.map((useCase) => (
              <div
                key={useCase.title}
                className="rounded-xl border border-navy-200 bg-navy-50/50 p-6"
              >
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-navy-900">
                    <useCase.icon className="h-5 w-5 text-white" />
                  </div>
                  <h3 className="text-lg font-semibold text-navy-900">
                    {useCase.title}
                  </h3>
                </div>
                <p className="mt-4 text-sm italic text-navy-600">
                  {useCase.example}
                </p>
                <p className="mt-3 text-sm leading-6 text-navy-500">
                  {useCase.output}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ===== PRICING ===== */}
      <section id="pricing" className="bg-navy-50 px-6 py-24 sm:py-32">
        <div className="mx-auto max-w-5xl">
          <div className="text-center">
            <h2 className="text-3xl font-bold tracking-tight text-navy-900 sm:text-4xl">
              Simple, transparent pricing
            </h2>
            <p className="mt-4 text-lg text-navy-500">
              Start free. Upgrade when you need more power.
            </p>
          </div>

          <div className="mt-16 grid grid-cols-1 gap-8 md:grid-cols-3">
            {pricingTiers.map((tier) => (
              <div
                key={tier.name}
                className={`relative flex flex-col rounded-2xl border p-8 ${
                  tier.highlighted
                    ? "border-brand bg-white shadow-xl shadow-brand/10 ring-1 ring-brand md:-mt-4 md:mb-[-1rem]"
                    : "border-navy-200 bg-white"
                }`}
              >
                {tier.badge && (
                  <div className="absolute -top-3.5 left-1/2 -translate-x-1/2">
                    <span className="rounded-full bg-brand px-4 py-1 text-xs font-semibold text-white">
                      {tier.badge}
                    </span>
                  </div>
                )}

                <div>
                  <h3 className="text-lg font-semibold text-navy-900">
                    {tier.name}
                  </h3>
                  <p className="mt-1 text-sm text-navy-500">
                    {tier.description}
                  </p>
                </div>

                <div className="mt-6 flex items-baseline gap-1">
                  <span className="text-4xl font-bold text-navy-900">
                    {tier.price}
                  </span>
                  <span className="text-sm text-navy-500">{tier.period}</span>
                </div>

                <ul className="mt-8 flex-1 space-y-3">
                  {tier.features.map((feature) => (
                    <li
                      key={feature}
                      className="flex items-start gap-3 text-sm text-navy-700"
                    >
                      <Check className="mt-0.5 h-4 w-4 shrink-0 text-brand" />
                      {feature}
                    </li>
                  ))}
                </ul>

                <Link
                  href={tier.href}
                  className={`mt-8 block rounded-lg px-4 py-3 text-center text-sm font-semibold transition-colors ${
                    tier.highlighted
                      ? "bg-brand text-white shadow-lg shadow-brand/25 hover:bg-brand-dark"
                      : "bg-navy-900 text-white hover:bg-navy-800"
                  }`}
                >
                  {tier.cta}
                </Link>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ===== CTA BANNER ===== */}
      <section className="bg-navy-950 px-6 py-24 sm:py-32">
        <div className="mx-auto max-w-3xl text-center">
          <h2 className="text-3xl font-bold tracking-tight text-white sm:text-4xl">
            Ready to plan your next move?
          </h2>
          <p className="mt-4 text-lg text-navy-300">
            Join thousands of founders and teams using GoPlan to turn ideas into
            action.
          </p>
          <Link
            href="/register"
            className="mt-10 inline-flex items-center gap-2 rounded-lg bg-brand px-8 py-3.5 text-base font-semibold text-white shadow-lg shadow-brand/25 transition-all hover:bg-brand-dark hover:shadow-xl hover:shadow-brand/30"
          >
            Start Planning — It&apos;s Free
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
      </section>

      {/* ===== FOOTER ===== */}
      <footer className="border-t border-navy-200 bg-white px-6 py-12">
        <div className="mx-auto flex max-w-5xl flex-col items-center justify-between gap-4 sm:flex-row">
          <p className="text-sm text-navy-500">
            &copy; 2026 GoPlan. All rights reserved.
          </p>
          <div className="flex gap-6">
            <Link
              href="#"
              className="text-sm text-navy-400 transition-colors hover:text-navy-700"
            >
              Privacy
            </Link>
            <Link
              href="#"
              className="text-sm text-navy-400 transition-colors hover:text-navy-700"
            >
              Terms
            </Link>
          </div>
        </div>
      </footer>
    </main>
  );
}
