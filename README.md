# littlecompiler


## Specification

```
func true() u8
    return u8(1)
end

func false() u8
    return u8(0)
end

func lang_spec_long_func(
        a i64,
        b i64,
        c u8) u8
    return u8(a) +
        u8(b) + c
end

func lang_spec()
    let a u8
    let b u16
    let c u32
    let d u64

    let e i8
    let f i16
    let g i32
    let h i64

    if true()
        let a u8

        a = u8(255) + u8(1) # 0
        a = u8(255) + u8(2) # 1

        a = u8(1) - u8(3) # 254

        a = u8(83) * u8(89) # 219
    end

    if true()
        let a i8

        a = i8(5) / i8(3) # 1
        a = i8(-5) / i8(3) # -1
        a = i8(5) / i8(-3) # -1
        a = i8(-5) / i8(-3) # 1

        a = i8(5) % i8(3) # 2
        a = i8(-5) % i8(3) # -2
        a = i8(5) % i8(-3) # 2
        a = i8(-5) % i8(-3) # -2

        a = i8(-128) / i8(-1) # -128
        a = i8(-128) % i8(-1) # 0

        a = i8(1) / i8(0) # PANIC
        a = i8(1) % i8(0) # PANIC
    end

    if true()
        let a u8
        let b i8

        a = u8(128) >> u8(1) # 64
        b = i8(-128) >> u8(1) # -64

        a = u8(128) >> u8(11) # 0
        b = i8(-128) >> u8(11) # -1

        a = u8(128) >> i8(1) # 64
        a = u8(128) >> i8(-1) # PANIC
        a = u8(128) << i8(-1) # PANIC

        a = u8(64) << i8(1) # 128
        a = u8(64) << i8(2) # 0
        b = i8(64) << i8(2) # 0
    end

    if true()
        let a i32
        let b i32

        a = i32(1)
        b = i32(2)

        if a == b
            a = i32(6)
        else if a == i32(0)
            a = i32(8)
        else
            b = i32(8)
        end

        if a != b
        end

        if a < b
        end

        if a <= b
        end

        if a > b
        end

        if a >= b
        end

        if (a >= b) && (a == i32(2))
        end

        if (a >= b) || (a == i32(2))
        end

        if (a >= b) ||
            (a == i32(2))
        end
    end

    if true()
        let a i32
        a = i32(10)
        while a > i32(0)
            a = a - i32(1)
        end
    end

    if true()
        let addr u64
        addr <- "ABC\"\\"
    end

    if true()
        let addr u64

        su8(addr, u8(1))
        su16(addr, u16(1))
        su32(addr, u32(1))
        su64(addr, u64(1))

        si8(addr, i8(1))
        si16(addr, i16(1))
        si32(addr, i32(1))
        si64(addr, i64(1))

        a = lu8(addr)
        b = lu16(addr)
        c = lu32(addr)
        d = lu64(addr)

        e = li8(addr)
        f = li16(addr)
        g = li32(addr)
        h = li64(addr)
    end

    if true()
        let a u8

        a = u8('A')
        a = u8('\'')
        a = u8('\\')
    end
end

func main()
    lang_spec()
end
```
