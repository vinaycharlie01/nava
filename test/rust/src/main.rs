fn add(a: i32, b: i32) -> i32 {
    a + b
}

fn main() {
    println!("sum={}.", add(2, 3));
}

#[cfg(test)]
mod tests {
    use super::add;

    #[test]
    fn adds_numbers() {
        assert_eq!(add(2, 3), 5);
    }
}
