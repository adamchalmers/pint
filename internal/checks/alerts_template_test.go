package checks_test

import (
	"testing"

	"github.com/cloudflare/pint/internal/checks"
	"github.com/cloudflare/pint/internal/promapi"
)

func newTemplateCheck(_ *promapi.FailoverGroup) checks.RuleChecker {
	return checks.NewTemplateCheck()
}

func TestTemplateCheck(t *testing.T) {
	testCases := []checkTest{
		{
			description: "skips recording rule",
			content:     "- record: foo\n  expr: sum(foo)\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems:    noProblems,
		},
		{
			description: "invalid syntax in annotations",
			content:     "- alert: Foo Is Down\n  expr: up{job=\"foo\"} == 0\n  annotations:\n    summary: 'Instance {{ $label.instance }} down'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `summary: Instance {{ $label.instance }} down`,
						Lines:    []int{4},
						Reporter: checks.TemplateCheckName,
						Text:     "template parse error: undefined variable \"$label\"",
						Severity: checks.Fatal,
					},
				}
			},
		},
		{
			description: "invalid function in annotations",
			content:     "- alert: Foo Is Down\n  expr: up{job=\"foo\"} == 0\n  annotations:\n    summary: '{{ $value | xxx }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `summary: {{ $value | xxx }}`,
						Lines:    []int{4},
						Reporter: checks.TemplateCheckName,
						Text:     "template parse error: function \"xxx\" not defined",
						Severity: checks.Fatal,
					},
				}
			},
		},
		{
			description: "valid syntax in annotations",
			content:     "- alert: Foo Is Down\n  expr: up{job=\"foo\"} == 0\n  annotations:\n    summary: 'Instance {{ $labels.instance }} down'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems:    noProblems,
		},
		{
			description: "invalid syntax in labels",
			content:     "- alert: Foo Is Down\n  expr: up{job=\"foo\"} == 0\n  labels:\n    summary: 'Instance {{ $label.instance }} down'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `summary: Instance {{ $label.instance }} down`,
						Lines:    []int{4},
						Reporter: checks.TemplateCheckName,
						Text:     "template parse error: undefined variable \"$label\"",
						Severity: checks.Fatal,
					},
				}
			},
		},
		{
			description: "invalid function in annotations",
			content:     "- alert: Foo Is Down\n  expr: up{job=\"foo\"} == 0\n  labels:\n    summary: '{{ $value | xxx }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `summary: {{ $value | xxx }}`,
						Lines:    []int{4},
						Reporter: checks.TemplateCheckName,
						Text:     "template parse error: function \"xxx\" not defined",
						Severity: checks.Fatal,
					},
				}
			},
		},
		{
			description: "valid syntax in labels",
			content:     "- alert: Foo Is Down\n  expr: up{job=\"foo\"} == 0\n  labels:\n    summary: 'Instance {{ $labels.instance }} down'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems:    noProblems,
		},
		{
			description: "{{ $value}} in label key",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    '{{ $value}}': bar\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: "{{ $value}}: bar",
						Lines:    []int{5},
						Reporter: checks.TemplateCheckName,
						Text:     "using $value in labels will generate a new alert on every value change, move it to annotations",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{ $value }} in label key",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    '{{ $value }}': bar\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: "{{ $value }}: bar",
						Lines:    []int{5},
						Reporter: checks.TemplateCheckName,
						Text:     "using $value in labels will generate a new alert on every value change, move it to annotations",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{$value}} in label value",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    baz: '{{$value}}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: "baz: {{$value}}",
						Lines:    []int{5},
						Reporter: checks.TemplateCheckName,
						Text:     "using $value in labels will generate a new alert on every value change, move it to annotations",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{$value}} in multiple labels",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: '{{ .Value }}'\n    baz: '{{$value}}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: "foo: {{ .Value }}",
						Lines:    []int{4},
						Reporter: checks.TemplateCheckName,
						Text:     "using .Value in labels will generate a new alert on every value change, move it to annotations",
						Severity: checks.Bug,
					},
					{
						Fragment: "baz: {{$value}}",
						Lines:    []int{5},
						Reporter: checks.TemplateCheckName,
						Text:     "using $value in labels will generate a new alert on every value change, move it to annotations",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{  $value  }} in label value",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    baz: |\n      foo is {{  $value | humanizePercentage }}%\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: "baz: foo is {{  $value | humanizePercentage }}%\n",
						Lines:    []int{5, 6},
						Reporter: checks.TemplateCheckName,
						Text:     "using $value in labels will generate a new alert on every value change, move it to annotations",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{  $value  }} in label value",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    baz: |\n      foo is {{$value|humanizePercentage}}%\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: "baz: foo is {{$value|humanizePercentage}}%\n",
						Lines:    []int{5, 6},
						Reporter: checks.TemplateCheckName,
						Text:     "using $value in labels will generate a new alert on every value change, move it to annotations",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{ .Value }} in label value",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    baz: 'value {{ .Value }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: "baz: value {{ .Value }}",
						Lines:    []int{5},
						Reporter: checks.TemplateCheckName,
						Text:     "using .Value in labels will generate a new alert on every value change, move it to annotations",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{ .Value|humanize }} in label value",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    baz: '{{ .Value|humanize }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: "baz: {{ .Value|humanize }}",
						Lines:    []int{5},
						Reporter: checks.TemplateCheckName,
						Text:     "using .Value in labels will generate a new alert on every value change, move it to annotations",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{ $foo := $value }} in label value",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    baz: '{{ $foo := $value }}{{ $foo }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: "baz: {{ $foo := $value }}{{ $foo }}",
						Lines:    []int{5},
						Reporter: checks.TemplateCheckName,
						Text:     "using $foo in labels will generate a new alert on every value change, move it to annotations",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "{{ $foo := .Value }} in label value",
			content:     "- alert: foo\n  expr: sum(foo)\n  labels:\n    foo: bar\n    baz: '{{ $foo := .Value }}{{ $foo }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: "baz: {{ $foo := .Value }}{{ $foo }}",
						Lines:    []int{5},
						Reporter: checks.TemplateCheckName,
						Text:     "using $foo in labels will generate a new alert on every value change, move it to annotations",
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (by)",
			content:     "- alert: Foo Is Down\n  expr: sum(foo) > 0\n  annotations:\n    summary: '{{ $labels.job }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `summary: {{ $labels.job }}`,
						Lines:    []int{2, 4},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "job" label but the query removes it`,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (by)",
			content:     "- alert: Foo Is Down\n  expr: sum(foo) > 0\n  annotations:\n    summary: '{{ .Labels.job }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `summary: {{ .Labels.job }}`,
						Lines:    []int{2, 4},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "job" label but the query removes it`,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (without)",
			content:     "- alert: Foo Is Down\n  expr: sum(foo) without(job) > 0\n  annotations:\n    summary: '{{ $labels.job }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `summary: {{ $labels.job }}`,
						Lines:    []int{2, 4},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "job" label but the query removes it`,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (without)",
			content:     "- alert: Foo Is Down\n  expr: sum(foo) without(job) > 0\n  annotations:\n    summary: '{{ .Labels.job }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `summary: {{ .Labels.job }}`,
						Lines:    []int{2, 4},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "job" label but the query removes it`,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "label missing from metrics (without)",
			content:     "- alert: Foo Is Down\n  expr: sum(foo) without(job) > 0\n  labels:\n    summary: '{{ $labels.job }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `summary: {{ $labels.job }}`,
						Lines:    []int{2, 4},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "job" label but the query removes it`,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (or)",
			content:     "- alert: Foo Is Down\n  expr: sum(foo) by(job) or sum(bar)\n  annotations:\n    summary: '{{ .Labels.job }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `summary: {{ .Labels.job }}`,
						Lines:    []int{2, 4},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "job" label but the query removes it`,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (1+)",
			content:     "- alert: Foo Is Down\n  expr: 1 + sum(foo) by(job) + sum(foo) by(notjob)\n  annotations:\n    summary: '{{ .Labels.job }}'\n",
			checker:     newTemplateCheck,
			prometheus:  noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `summary: {{ .Labels.job }}`,
						Lines:    []int{2, 4},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "job" label but the query removes it`,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (group_left)",
			content: `
- alert: Foo Is Down
  expr: count(build_info) by (instance, version) != ignoring(package) group_left(foo) count(package_installed) by (instance, version, package)
  annotations:
    summary: '{{ $labels.instance }} on {{ .Labels.foo }} is down'
    help: '{{ $labels.ixtance }}'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `help: {{ $labels.ixtance }}`,
						Lines:    []int{3, 6},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "ixtance" label but the query removes it`,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label present on metrics (absent)",
			content: `
- alert: Foo Is Missing
  expr: absent(foo{job="bar", instance="server1"})
  annotations:
    summary: '{{ $labels.instance }} on {{ .Labels.job }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "annotation label missing from metrics (absent)",
			content: `
- alert: Foo Is Missing
  expr: absent(foo{job="bar"}) AND on(job) foo
  labels:
    instance: '{{ $labels.instance }}'
  annotations:
    summary: '{{ $labels.instance }} on {{ .Labels.foo }} is missing'
    help: '{{ $labels.xxx }}'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: "instance: {{ $labels.instance }}",
						Lines:    []int{3, 5},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "instance" label but absent() is not passing it`,
						Severity: checks.Bug,
					},
					{
						Fragment: `summary: {{ $labels.instance }} on {{ .Labels.foo }} is missing`,
						Lines:    []int{3, 7},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "instance" label but absent() is not passing it`,
						Severity: checks.Bug,
					},
					{
						Fragment: `summary: {{ $labels.instance }} on {{ .Labels.foo }} is missing`,
						Lines:    []int{3, 7},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "foo" label but absent() is not passing it`,
						Severity: checks.Bug,
					},
					{
						Fragment: "help: {{ $labels.xxx }}",
						Lines:    []int{3, 8},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "xxx" label but absent() is not passing it`,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label present on metrics (absent(sum))",
			content: `
- alert: Foo Is Missing
  expr: absent(sum(foo) by(job, instance))
  annotations:
    summary: '{{ $labels.instance }} on {{ .Labels.job }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "annotation label missing from metrics (absent(sum))",
			content: `
- alert: Foo Is Missing
  expr: absent(sum(foo) by(job))
  annotations:
    summary: '{{ $labels.instance }} on {{ .Labels.job }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `summary: {{ $labels.instance }} on {{ .Labels.job }} is missing`,
						Lines:    []int{3, 5},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "instance" label but the query removes it`,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (absent({job=~}))",
			content: `
- alert: Foo Is Missing
  expr: absent({job=~".+"})
  annotations:
    summary: '{{ .Labels.job }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `summary: {{ .Labels.job }} is missing`,
						Lines:    []int{3, 5},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "job" label but absent() is not passing it`,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "annotation label missing from metrics (absent()) / multiple",
			content: `
- alert: Foo Is Missing
  expr: absent(foo) or absent(bar)
  annotations:
    summary: '{{ .Labels.job }} / {{$labels.job}} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems: func(uri string) []checks.Problem {
				return []checks.Problem{
					{
						Fragment: `summary: {{ .Labels.job }} / {{$labels.job}} is missing`,
						Lines:    []int{3, 5},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "job" label but absent() is not passing it`,
						Severity: checks.Bug,
					},
					{
						Fragment: `summary: {{ .Labels.job }} / {{$labels.job}} is missing`,
						Lines:    []int{3, 5},
						Reporter: checks.TemplateCheckName,
						Text:     `template is using "job" label but absent() is not passing it`,
						Severity: checks.Bug,
					},
				}
			},
		},
		{
			description: "absent() * on() group_left(...) foo",
			content: `
- alert: Foo
  expr: absent(foo{job="xxx"}) * on() group_left(cluster, env) bar
  annotations:
    summary: '{{ .Labels.job }} in cluster {{$labels.cluster}}/{{ $labels.env }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "absent() * on() group_left() bar",
			content: `
- alert: Foo
  expr: absent(foo{job="xxx"}) * on() group_left() bar
  annotations:
    summary: '{{ .Labels.job }} in cluster {{$labels.cluster}}/{{ $labels.env }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "bar * on() group_right(...) absent()",
			content: `
- alert: Foo
  expr: bar * on() group_right(cluster, env) absent(foo{job="xxx"})
  annotations:
    summary: '{{ .Labels.job }} in cluster {{$labels.cluster}}/{{ $labels.env }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "bar * on() group_right() absent()",
			content: `
- alert: Foo
  expr: bar * on() group_right() absent(foo{job="xxx"})
  annotations:
    summary: '{{ .Labels.job }} in cluster {{$labels.cluster}}/{{ $labels.env }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
		{
			description: "",
			content: `
- alert: Foo
  expr: foo and on() absent(bar)
  annotations:
    summary: '{{ .Labels.job }} is missing'
`,
			checker:    newTemplateCheck,
			prometheus: noProm,
			problems:   noProblems,
		},
	}
	runTests(t, testCases)
}
