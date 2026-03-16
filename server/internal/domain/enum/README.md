# internal/domain/enum

## 作用
集中维护枚举常量，避免字符串散落。

## V1 枚举范围
- 角色：super_admin / operator。
- 车型状态：draft / published / unpublished / deleted。
- 价格模式：show_price / phone_inquiry。
- 资源类型：image / video。

## 实践建议
- 枚举需有合法性校验函数。
- 枚举变更必须同步影响范围清单。
