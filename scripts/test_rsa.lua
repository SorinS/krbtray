 -- Prompt for RSA PIN
  local pin, ok = ktray.prompt_secret("RSA Authentication", "Please enter your RSA PIN:")
  if not ok then
      ktray.set_status("Cancelled")
      return
  end

  ktray.set_status("RSA Received")

  -- Use the PIN with Kerberos token
  local token,ok = ktray.get_token("BAM")
  if not ok then
     ktray.set_status("error getting the krb token for BAM")
     return
  end

  local headers = {
      ["Authorization"] = "Negotiate " .. token,
      ["X-RSA-Token"] = pin
  }
  ktray.set_status("Kerberos token for BAM received")
