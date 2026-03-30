# Philosophy Audit

## Overview

提供审计 SPECS 文档的清单和方法，验证规范是否引用哲学依据、规范是否与哲学一致、哲学覆盖度是否完整。

定义什么：哲学审计的检查项、审计方法、修复指南。不涵盖：具体哲学内容（由 02-golang-philosophy.md、03-typescript-philosophy.md 定义）。

相关 SPEC：[00-principles.md]（核心原则）、[02-golang-philosophy.md]（Golang 哲学）、[03-typescript-philosophy.md]（TypeScript 哲学）。

## 核心原则

### Philosophy-First 强制要求

所有 SPECS 都是 philosophy 思想的具体体现。每条规范必须能回答"这条规则体现了哪个哲学原则"。

如果无法回答：
- 规则本身不必要 → 删除规则
- 哲学文档不完整 → 补充哲学

### 100% 可追溯性

每条 SPEC 规则必须包含 `> **Philosophy**` 块，说明其体现的哲学原则。

## 审计清单

### 1. 哲学引用完整性检查

检查 SPEC 文档是否包含哲学引用。

**检查项**：
- [ ] 每个规范章节包含 `> **Philosophy**` 块
- [ ] Philosophy 块引用具体哲学文档和章节（格式：`见 XX.md §YYY`）
- [ ] Philosophy 块说明如何体现哲学原则（不只是引用）
- [ ] 引用的哲学文档存在且可访问
- [ ] 引用的章节在哲学文档中存在

**检查方法**：
```bash
# 检查是否包含 Philosophy 块
grep -r "> \*\*Philosophy\*\*" SPECS/40-*.md SPECS/41-*.md SPECS/42-*.md SPECS/43-*.md

# 检查引用格式
grep -r "见 [0-9][0-9]-.*\.md §" SPECS/
```

**修复指南**：
- 缺少 Philosophy 块 → 添加 `> **Philosophy**: 体现 XXX 原则（见 YY.md §ZZZ）。[说明如何体现]`
- 引用格式错误 → 使用标准格式：`见 02-golang-philosophy.md §Clear is better than clever`
- 引用不存在的文档 → 检查文档路径，或补充哲学文档

### 2. 哲学一致性检查

检查规范是否与引用的哲学原则一致。

**检查项**：
- [ ] 规范的要求与哲学原则不矛盾
- [ ] 规范的示例体现哲学原则
- [ ] 规范的 Forbidden 章节与哲学一致
- [ ] 规范的决策依据（`> **Why**`）与哲学对齐

**检查方法**：
1. 读取规范中的 Philosophy 引用
2. 读取对应的哲学原则
3. 验证规范是否体现该原则

**示例**：

✅ **一致**：
```markdown
### 函数长度限制
函数不超过 50 行（不含注释和空行）。

> **Philosophy**: 体现 KISS 原则（见 02-golang-philosophy.md §KISS）和单一职责原则（见 02-golang-philosophy.md §Single Responsibility Principle）。短函数更易理解、测试和维护。
> **Why**: 超过 50 行的函数通常承担多个职责，违反单一职责原则。
```

❌ **不一致**：
```markdown
### 函数长度限制
函数不超过 200 行。

> **Philosophy**: 体现 KISS 原则（见 02-golang-philosophy.md §KISS）。
> **Why**: 长函数更灵活。
```
问题：200 行不符合 KISS 原则，"更灵活"与 KISS 矛盾。

**修复指南**：
- 规范与哲学矛盾 → 修改规范以符合哲学，或更新哲学文档
- 示例不体现哲学 → 更新示例，添加 Good/Bad 对比
- Why 与哲学不对齐 → 重写 Why，明确说明如何体现哲学

### 3. 哲学覆盖度检查

检查哲学文档是否覆盖所有核心决策。

**检查项**：
- [ ] 所有编码规范都能追溯到哲学
- [ ] 所有架构规范都能追溯到哲学
- [ ] 所有 API 规范都能追溯到哲学
- [ ] 所有数据模型规范都能追溯到哲学
- [ ] 哲学文档包含 KISS、DRY、YAGNI、SOLID 原则
- [ ] 哲学文档包含技术栈特定原则（Go Proverbs、TypeScript Trust）

**检查方法**：
1. 列出所有 SPEC 规范
2. 检查每条规范是否有 Philosophy 引用
3. 统计覆盖率

```bash
# 统计 SPEC 规范总数
grep -c "^### " SPECS/40-*.md SPECS/41-*.md SPECS/42-*.md SPECS/43-*.md

# 统计包含 Philosophy 引用的规范数
grep -c "> \*\*Philosophy\*\*" SPECS/40-*.md SPECS/41-*.md SPECS/42-*.md SPECS/43-*.md

# 计算覆盖率
```

**目标覆盖率**：100%

**修复指南**：
- 覆盖率 < 100% → 为缺失的规范添加 Philosophy 引用
- 规范无法追溯到哲学 → 补充哲学文档或删除规范
- 哲学文档缺少核心原则 → 补充哲学原则

### 4. 哲学文档质量检查

检查哲学文档本身的质量。

**检查项**：
- [ ] 每个原则有 Core Thesis（1-3 句话）
- [ ] 每个原则有详细解释（2-3 段）
- [ ] 每个原则有正反示例（Good/Bad）
- [ ] 每个原则有 `> **Why**` 说明
- [ ] 每个原则有 `> **Rejected**` 说明
- [ ] 哲学文档包含 Terminology 章节
- [ ] 哲学文档包含 Forbidden 章节
- [ ] 哲学文档包含 References 章节

**检查方法**：
```bash
# 检查哲学文档结构
grep "^## Core Thesis" SPECS/*-philosophy*.md
grep "^## Principles" SPECS/*-philosophy*.md
grep "^## Terminology" SPECS/*-philosophy*.md
grep "^## Forbidden" SPECS/*-philosophy*.md

# 检查原则是否有示例
grep -A 20 "^### " SPECS/*-philosophy*.md | grep "好的做法"
grep -A 20 "^### " SPECS/*-philosophy*.md | grep "不好的做法"

# 检查原则是否有 Why
grep -A 20 "^### " SPECS/*-philosophy*.md | grep "> \*\*Why\*\*"
```

**修复指南**：
- 缺少 Core Thesis → 添加 1-3 句话概括哲学
- 缺少示例 → 添加 Good/Bad 代码示例
- 缺少 Why → 添加 `> **Why**` 说明原则重要性
- 缺少 Rejected → 添加 `> **Rejected**` 说明拒绝的方案

### 5. 引用格式一致性检查

检查 Philosophy 引用格式是否统一。

**标准格式**：
```markdown
> **Philosophy**: 体现 XXX 原则（见 YY.md §ZZZ）。[说明如何体现该原则]
```

**检查项**：
- [ ] 使用 `> **Philosophy**:` 开头（注意冒号）
- [ ] 使用"体现"而非"遵循"或"符合"
- [ ] 引用格式为 `（见 XX.md §YYY）`
- [ ] 包含说明如何体现原则的文字

**检查方法**：
```bash
# 检查格式
grep "> \*\*Philosophy\*\*:" SPECS/40-*.md SPECS/41-*.md SPECS/42-*.md SPECS/43-*.md

# 检查是否包含"体现"
grep "> \*\*Philosophy\*\*: 体现" SPECS/40-*.md SPECS/41-*.md SPECS/42-*.md SPECS/43-*.md
```

**修复指南**：
- 格式不统一 → 使用标准格式
- 缺少说明 → 添加如何体现原则的说明
- 引用格式错误 → 修正为 `（见 XX.md §YYY）`

## 审计流程

### 阶段 1：自动化检查

使用脚本自动检查基本项：

```bash
#!/bin/bash
# philosophy-audit.sh

echo "=== Philosophy Audit ==="

# 1. 检查哲学文档是否存在
echo "1. Checking philosophy documents..."
for file in SPECS/02-golang-philosophy.md SPECS/03-typescript-philosophy.md; do
  if [ -f "$file" ]; then
    echo "  ✓ $file exists"
  else
    echo "  ✗ $file missing"
  fi
done

# 2. 检查 SPEC 是否包含 Philosophy 引用
echo "2. Checking Philosophy references..."
for file in SPECS/40-*.md SPECS/41-*.md SPECS/42-*.md SPECS/43-*.md; do
  if [ -f "$file" ]; then
    count=$(grep -c "> \*\*Philosophy\*\*" "$file" || echo 0)
    echo "  $file: $count references"
  fi
done

# 3. 检查引用格式
echo "3. Checking reference format..."
grep -n "> \*\*Philosophy\*\*" SPECS/40-*.md SPECS/41-*.md SPECS/42-*.md SPECS/43-*.md | \
  grep -v "见 [0-9][0-9]-.*\.md §" && \
  echo "  ✗ Found invalid reference format" || \
  echo "  ✓ All references use correct format"

# 4. 检查哲学文档结构
echo "4. Checking philosophy document structure..."
for file in SPECS/*-philosophy*.md; do
  if [ -f "$file" ]; then
    echo "  Checking $file..."
    grep -q "^## Core Thesis" "$file" && echo "    ✓ Has Core Thesis" || echo "    ✗ Missing Core Thesis"
    grep -q "^## Principles" "$file" && echo "    ✓ Has Principles" || echo "    ✗ Missing Principles"
    grep -q "^## Terminology" "$file" && echo "    ✓ Has Terminology" || echo "    ✗ Missing Terminology"
    grep -q "^## Forbidden" "$file" && echo "    ✓ Has Forbidden" || echo "    ✗ Missing Forbidden"
  fi
done

echo "=== Audit Complete ==="
```

### 阶段 2：人工审查

自动化检查后，进行人工审查：

1. **一致性审查**：读取规范和对应哲学，验证是否一致
2. **质量审查**：检查哲学原则是否有足够的解释和示例
3. **覆盖度审查**：检查是否所有核心决策都有哲学支撑

### 阶段 3：修复和验证

根据审计结果修复问题，然后重新审计验证。

## 审计报告模板

```markdown
# Philosophy Audit Report

**Date**: YYYY-MM-DD
**Auditor**: [Name]

## Summary

- Total SPEC files audited: X
- Total rules audited: Y
- Rules with Philosophy references: Z
- Coverage rate: Z/Y = XX%

## Findings

### Critical Issues
- [ ] Issue 1: Description
- [ ] Issue 2: Description

### Major Issues
- [ ] Issue 1: Description
- [ ] Issue 2: Description

### Minor Issues
- [ ] Issue 1: Description
- [ ] Issue 2: Description

## Recommendations

1. Recommendation 1
2. Recommendation 2
3. Recommendation 3

## Action Items

- [ ] Action 1: Owner, Deadline
- [ ] Action 2: Owner, Deadline
- [ ] Action 3: Owner, Deadline
```

## 常见问题

### Q1: 规范无法追溯到现有哲学原则怎么办？

**A**: 两种选择：
1. 补充哲学文档，添加缺失的原则
2. 删除规范，如果该规范不必要

### Q2: 一条规范体现多个哲学原则怎么办？

**A**: 在 Philosophy 块中列出所有相关原则：
```markdown
> **Philosophy**: 体现 KISS 原则（见 02-golang-philosophy.md §KISS）和单一职责原则（见 02-golang-philosophy.md §Single Responsibility Principle）。短函数更易理解、测试和维护。
```

### Q3: 哲学原则之间有冲突怎么办？

**A**: 在哲学文档中使用 `> **Rejected**` 说明权衡：
```markdown
> **Why**: 清晰性优于性能。只有在性能瓶颈确认后才优化。
> **Rejected**: 过早优化 — 牺牲清晰性换取未经验证的性能提升。
```

### Q4: 审计频率应该是多少？

**A**:
- 新增 SPEC 时：立即审计
- 修改 SPEC 时：立即审计
- 定期审计：每季度一次
- 重大重构前：完整审计

## Philosophy Coverage Metrics

### Coverage Formula

```
Philosophy Coverage = (Rules with Philosophy blocks) / (Total rules) × 100%
```

### Target Coverage

| Scope | Target | Rationale |
|-------|--------|-----------|
| Meta-project SPECS | 80%+ | Serves as reference implementation |
| Project SPECS (40-43 series) | 60%+ | Critical rules must have philosophical backing |
| Project SPECS (other series) | 40%+ | Recommended but not enforced |

### Calculation

```bash
# Count rules with Philosophy blocks
philosophy_count=$(grep -r "> \*\*Philosophy\*\*" SPECS/ | wc -l | tr -d ' ')

# Count total rules (### headers in 40-43 series)
total_rules=$(grep -r "^### " SPECS/4[0-3]-*.md | wc -l | tr -d ' ')

# Calculate coverage
echo "Coverage: $philosophy_count / $total_rules"
```

## Terminology

| Term | Definition | Not |
|------|-----------|-----|
| **Philosophy Audit** | 审计 SPECS 是否引用哲学依据的过程 | Not "代码审查"（审计的是文档，不是代码） |
| **Coverage Rate** | 包含 Philosophy 引用的规范占总规范的比例 | Not "测试覆盖率"（这是文档覆盖率） |
| **Traceability** | 规范能追溯到哲学原则的能力 | Not "可追踪性"（这是设计决策的可追溯性） |

## Forbidden

- **Don't skip Philosophy references**: 每条规范必须有哲学引用 → 100% 覆盖率是强制要求
- **Don't use vague references**: "体现最佳实践"不是有效引用 → 必须引用具体哲学原则和章节
- **Don't ignore inconsistencies**: 规范与哲学矛盾必须修复 → 不能"先这样，以后再说"
- **Don't audit without fixing**: 审计发现问题必须修复 → 审计不是走形式

## References

- [00-principles.md] — SPEC 核心原则
- [02-golang-philosophy.md] — Golang 哲学
- [03-typescript-philosophy.md] — TypeScript 哲学
- [11-structure.md] — SPEC 结构规范
