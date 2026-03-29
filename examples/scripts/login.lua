function main()
    return {
        method = "POST",
        headers = {
            ["Content-Type"] = "application/json"
        },
        body = '{"username":"user_' .. __randomInt(1, 99999) .. '","id":"' .. __uuidv4() .. '"}'
    }
end