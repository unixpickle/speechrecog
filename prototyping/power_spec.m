function [spec] = power_spec(signal)
  fftRes = fft(signal);
  spec = zeros(rows(signal)/2+1, 1);
  for i = 1:(rows(signal)/2+1)
    spec(i) = fftRes(i)*conj(fftRes(i)) / rows(signal);
  end
end
